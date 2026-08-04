# Kubernetes AI Lab

> A self-hosted Kubernetes laboratory designed to automate the processing of university lectures using local AI models while serving as a hands-on environment to learn Kubernetes, GitOps, and cloud-native technologies.

---

# Overview

This project is both a learning platform and a real-world automation pipeline.

It was created to support my Master's degree studies by automatically transforming recorded lectures into structured study notes using entirely self-hosted AI.

The complete workflow runs inside a local Kubernetes cluster, avoiding any dependency on external AI APIs.

The pipeline performs the following tasks:

1. Store lecture recordings in Nextcloud.
2. Detect new audio files automatically.
3. Download them through WebDAV.
4. Transcribe the audio using Whisper.
5. Process the transcription using a local LLM running on Ollama.
6. Generate clean, structured Markdown notes.
7. Upload the generated notes back to Nextcloud.

Everything is fully automated.

---

# Architecture

```
                     +----------------+
                     |   Nextcloud    |
                     | Audio Storage  |
                     +--------+-------+
                              |
                              |
                              v
                  +----------------------+
                  |     Go Puller        |
                  |  WebDAV Scanner      |
                  +----------+-----------+
                             |
                             |
                             v
                     +---------------+
                     |    Whisper    |
                     | Speech-to-Text|
                     +-------+-------+
                             |
                             |
                             v
                     +---------------+
                     |    Ollama     |
                     |   Local LLM   |
                     +-------+-------+
                             |
                             |
                             v
                 Structured Markdown Notes
                             |
                             |
                             v
                     +---------------+
                     |   Nextcloud   |
                     +---------------+
```

---

# Technologies

## Kubernetes

The entire laboratory runs inside Kubernetes.

It provides a production-like environment while remaining completely local.

This project is used to learn:

- Deployments
- Services
- Ingress
- Pods
- Networking
- Resource Management
- Health Probes
- Persistent Volumes
- Multi-node scheduling

---

## Minikube

The Kubernetes cluster is powered by Minikube.

The cluster consists of multiple worker nodes, allowing experimentation with distributed workloads while remaining lightweight enough to run on a personal machine.

---

## ArgoCD

The whole infrastructure follows a GitOps workflow.

Every Kubernetes resource is stored inside this repository.

ArgoCD continuously synchronizes the cluster with Git, making Git the single source of truth.

Benefits include:

- Infrastructure as Code
- Automatic synchronization
- Easy recovery
- Version control
- Declarative deployments

---

## Tailscale

Remote access is provided through Tailscale.

Instead of exposing services to the public Internet, every application is securely accessible only through the private Tailnet.

This includes:

- Nextcloud
- Whisper
- Ollama
- ArgoCD

All communications are encrypted.

---

# Components

## Nextcloud

Nextcloud acts as the central storage system.

It stores:

- Lecture recordings
- Documents
- PDFs
- Generated Markdown notes
- Study materials

The Go service interacts with Nextcloud using the WebDAV protocol.

---

## Go Puller

The Puller is a custom application written in Go.

It periodically scans the complete Nextcloud directory tree looking for new audio files.

Responsibilities include:

- Discover folders
- List files
- Detect supported audio formats
- Download audio files
- Send audio to Whisper
- Send transcriptions to Ollama
- Upload generated Markdown notes back to Nextcloud

This service orchestrates the complete workflow.

---

## Whisper

Whisper provides speech-to-text capabilities.

It runs as a FastAPI application inside Kubernetes and exposes a simple REST endpoint.

```
POST /transcribe
```

The endpoint accepts an audio file using multipart/form-data and returns a JSON response similar to:

```json
{
  "language": "en",
  "text": "Complete transcription..."
}
```

The service uses Faster-Whisper running entirely on CPU.

No external APIs are required.

---

## Ollama

Once Whisper generates the transcription, the text is forwarded to Ollama.

Ollama hosts local Large Language Models capable of:

- Cleaning noisy transcriptions
- Correcting grammar
- Removing repetitions
- Summarizing lectures
- Organizing information
- Generating study notes
- Creating structured Markdown
- Producing outlines and key concepts

Everything runs locally.

---

# Complete Workflow

```
Lecture Recording
        |
        v
   Nextcloud Storage
        |
        v
 Go Puller detects audio
        |
        v
 Download via WebDAV
        |
        v
     Whisper API
        |
        v
   Speech-to-Text
        |
        v
      Ollama
        |
        v
 Markdown Generation
        |
        v
 Upload to Nextcloud
```

---

# Repository Structure

```
.
├── argocd/
├── deployments/
├── ingress/
├── nextcloud/
├── ollama/
├── puller/
├── services/
├── whisper/
└── README.md
```

---

# Features

- Kubernetes
- Minikube
- Multi-node Cluster
- Docker
- GitOps
- ArgoCD
- Tailscale
- Nextcloud
- WebDAV
- Whisper
- Ollama
- FastAPI
- Go
- Local AI
- Markdown Generation
- Fully Self-Hosted

---

# Learning Goals

This laboratory was created to gain practical experience with:

- Kubernetes Administration
- GitOps Workflows
- Containerization
- Docker
- Kubernetes Networking
- Service Discovery
- Ingress Controllers
- ArgoCD
- WebDAV
- REST APIs
- Go Development
- FastAPI
- Local AI Models
- Workflow Automation

Rather than building isolated examples, the objective is to create a complete system that solves a real problem while learning modern cloud-native technologies.

---

# Why Local AI?

One of the project's main goals is to avoid relying on external AI services.

Keeping every component local provides several advantages:

- Privacy
- No API costs
- Low latency
- Full control over the infrastructure
- Offline capabilities
- Complete ownership of the data

Every stage of the pipeline runs inside the Kubernetes cluster.

---

# Future Improvements

Planned enhancements include:

- Concurrent audio processing
- Job queues (RabbitMQ or NATS)
- Incremental Nextcloud synchronization
- Metadata database
- Automatic flashcard generation
- Obsidian integration
- Mind map generation
- RAG support
- Vector database integration
- Prometheus monitoring
- Grafana dashboards
- Helm Charts
- CI/CD pipelines
- GPU acceleration

---

# Project Status

🚧 **Work in Progress**

This laboratory continuously evolves as I progress through my Master's degree and explore new Kubernetes, DevOps, and AI technologies.

---

# License

This project is intended for educational purposes, experimentation, and continuous learning.