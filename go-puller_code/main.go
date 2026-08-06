package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	NEXTCLOUD_USERNAME = os.Getenv("NEXTCLOUD_USERNAME")
	NEXTCLOUD_PSW      = os.Getenv("NEXTCLOUD_PSW")
	NEXTCLOUD_DOMAIN   = os.Getenv("NEXT_CLOUD_DOMAIN")

	SCOPED_FOLDER_PATH = os.Getenv("SCOPED_FOLDER_PATH")

	OLLAMA_AI_ENDPOINT = os.Getenv("OLLAMA_AI_ENDPOINT")
	OLLAMA_AGENT_MODEL = os.Getenv("OLLAMA_AGENT_MODEL")
	OLLAMA_PROMPT_FILE = os.Getenv("OLLAMA_PROMPT_FILE")

	WHISPER_AI_ENDPOINT = os.Getenv("WHISPER_AI_ENDPOINT")
)

type MultiStatus struct {
	Responses []Response `xml:"response"`
}

type Response struct {
	Href     string     `xml:"href"`
	PropStat []PropStat `xml:"propstat"`
}

type PropStat struct {
	Prop Prop `xml:"prop"`
}

type Prop struct {
	LastModified string       `xml:"getlastmodified"`
	ContentType  string       `xml:"getcontenttype"`
	ContentSize  int64        `xml:"getcontentlength"`
	ResourceType ResourceType `xml:"resourcetype"`
}

type ResourceType struct {
	Collection *struct{} `xml:"collection"`
}

type WhisperResponse struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

func checkFolder(path string) {

	url := NEXTCLOUD_DOMAIN + path

	fmt.Println("funcion check_folder :" + url)

	req, err := http.NewRequest("PROPFIND", url, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)
	req.Header.Set("Depth", "1")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var folder MultiStatus

	err = xml.Unmarshal(body, &folder)
	if err != nil {
		panic(err)
	}

	for _, f := range folder.Responses {

		if len(f.PropStat) == 0 {
			continue
		}

		prop := f.PropStat[0].Prop
		isDir := prop.ResourceType.Collection != nil

		// Ignorar carpetas
		if isDir {
			continue
		}

		t, err := time.Parse(time.RFC1123, prop.LastModified)
		if err != nil {
			continue
		}

		if t.After(time.Now().Add(-5 * time.Minute)) {
			if check_is_audio(path) {
				fmt.Println("Is audio")
			}
		}
	}
}

func get_folder_paths() []string {
	var result MultiStatus
	var path_list []string

	url := SCOPED_FOLDER_PATH

	fmt.Println("fujncion get folder path: " + url)

	req, err := http.NewRequest("PROPFIND", url, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)
	req.Header.Set("Depth", "1")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	xml := xml.Unmarshal(body, &result)
	if xml != nil {
		log.Fatal(err)
	}

	for i, r := range result.Responses {

		if i == 0 {
			continue
		}

		if len(r.PropStat) == 0 {
			continue
		}

		prop := r.PropStat[0].Prop

		isDir := prop.ResourceType.Collection != nil

		if isDir {
			full_path := NEXTCLOUD_DOMAIN + r.Href + "Audios_toDo"
			path_list = append(path_list, full_path)
		}

	}

	return path_list
}

func get_files_from_path(path string) []string {
	var result MultiStatus
	var files []string
	url := path

	req, err := http.NewRequest("PROPFIND", url, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)
	req.Header.Set("Depth", "1")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	xml := xml.Unmarshal(body, &result)
	if xml != nil {
		log.Fatal(err)
	}

	if len(result.Responses) > 1 {
		for i, p := range result.Responses {

			if i == 0 {
				continue
			}

			files = append(files, p.Href)

		}
	}
	return files

}

func check_is_audio(path string) bool {
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	return strings.HasPrefix(mimeType, "audio/")
}

func sendToWhisper(audio io.Reader, filename string) (string, error) {

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, audio)
	if err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		WHISPER_AI_ENDPOINT,
		&body,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("whisper returned %s: %s", resp.Status, string(responseBody))
	}

	var result WhisperResponse

	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}

	return result.Text, nil
}

func sendToOllama(transcription string) (string, error) {

	type OllamaRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream"`
	}

	type OllamaResponse struct {
		Model    string `json:"model"`
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	fmt.Printf("Transcription: %q\n", transcription)

	promptBytes, err := os.ReadFile(OLLAMA_PROMPT_FILE)
	if err != nil {
		return "", err
	}

	prompt := string(promptBytes) +
		"\n\n---\n\n" +
		"CLASS TRANSCRIPT:\n\n" +
		transcription

	reqBody := OllamaRequest{
		Model:  OLLAMA_AGENT_MODEL,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	fmt.Println("========== PROMPT ==========")
	fmt.Println(prompt)
	fmt.Println("============================")

	req, err := http.NewRequest(
		"POST",
		OLLAMA_AI_ENDPOINT,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned %s: %s", resp.Status, string(body))
	}

	var result OllamaResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Response, nil
}

func saveMarkdown(markdown, markdownPath string) error {

	req, err := http.NewRequest("PUT", markdownPath, strings.NewReader(markdown))
	if err != nil {
		return err
	}

	req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)
	req.Header.Set("Content-Type", "text/markdown; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	return nil
}

func copyAudio(audioBytes []byte, destinationPath string) error {

	req, err := http.NewRequest("PUT", destinationPath, bytes.NewReader(audioBytes))
	if err != nil {
		return err
	}

	req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)
	req.Header.Set("Content-Type", "audio/wav")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	return nil
}

func deleteAudio(audioPath string) error {

	req, err := http.NewRequest("DELETE", audioPath, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	return nil
}

func main() {

	var path_list []string
	var file_list []string

	fmt.Println("hola")
	path_list = get_folder_paths()
	for _, folder_path := range path_list {
		file_list = get_files_from_path(folder_path)

		write_markdown_path := strings.Replace(folder_path, "Audios_toDo", "Audio_results", 1)
		copy_audio_path := strings.Replace(folder_path, "Audios_toDo", "Audios", 1)

		if len(file_list) != 0 {

			for _, v := range file_list {

				if check_is_audio(v) {
					fmt.Println("Es audio :" + v)

					audio_name := path.Base(v)
					fmt.Println(audio_name)

					url := NEXTCLOUD_DOMAIN + v

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						log.Fatal(err)
					}

					req.SetBasicAuth(NEXTCLOUD_USERNAME, NEXTCLOUD_PSW)

					client := &http.Client{}

					resp, err := client.Do(req)
					if err != nil {
						log.Fatal(err)
					}
					defer resp.Body.Close()

					audioBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Fatal(err)
					}

					resp.Body.Close()

					transcribedAudio, err := sendToWhisper(bytes.NewReader(audioBytes), audio_name+".wav")
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("=== Transcription ===")
					fmt.Println(transcribedAudio)

					markdown, err := sendToOllama(transcribedAudio)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("=== Generated Markdown ===")
					fmt.Println(markdown)

					saveMarkdown(markdown, write_markdown_path+"/"+audio_name+".md")
					copyAudio(audioBytes, copy_audio_path+"/"+audio_name+".wav")
					deleteAudio(url)
				}

			}
		}
	}
}
