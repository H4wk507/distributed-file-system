# Distributed File System

Rozproszony system plików z architekturą master-storage, zaimplementowany w Go z frontendem w React.

## Architektura

- **Master Node** – koordynuje klaster, zarządza metadanymi i distributed locks
- **Storage Nodes** – przechowują pliki z replikacją
- **API** – REST API do komunikacji z frontendem
- **Frontend** – React + Vite + TailwindCSS

Kluczowe mechanizmy:

- **Bully Election** – automatyczny wybór lidera przy awarii mastera
- **Lamport Timestamps** – synchronizacja zdarzeń w klastrze
- **Wait-for Graph** – detekcja i rozwiązywanie deadlocków
- **Consistent Hashing** – dystrybucja plików między węzłami
- **Streaming Protocol** – binarny protokół TCP do transferu plików

### Streaming Protocol

Dedykowany binarny protokół TCP do wydajnego przesyłania plików z chunked upload/download:

```
┌─────────────────────────────────────────────────────────────────┐
│                    STREAM HEADER (64 bytes)                     │
├────────┬──────────────┬───────────┬────────────┬───────────────┤
│ Magic  │  Session ID  │ File Size │ Chunk Size │   Reserved    │
│ 1 byte │   36 bytes   │  8 bytes  │  4 bytes   │   15 bytes    │
└────────┴──────────────┴───────────┴────────────┴───────────────┘
```

- **Magic byte**: `0x01` = upload, `0x02` = download
- **Chunk size**: domyślnie 4MB z buffer pooling
- **Streaming port**: port główny węzła + 100 (np. storage1:9001 → stream:9101)

## Uruchomienie lokalne

### Wymagania

- Docker (preferowana wersja 23+ dla wsparcia BuildKit) & Docker Compose
- Node.js 22.x + pnpm 10.x
- Go 1.21+ (opcjonalnie, do eksperymentów)
- Go migrate - do migracji bazy danych

### 1. Backend (Docker Compose)

```bash
docker compose up -d --build
```

Serwisy:

- API: http://localhost:8080
- Master: localhost:9000
- Storage nodes: localhost:9001-9003

### 2. Frontend

```bash
cd frontend
pnpm install
pnpm run dev
```

Frontend będzie dostępny pod http://localhost:5173

## Eksperymenty

Szczegóły w [experiments/README.md](experiments/README.md) – testy wydajności elekcji, locków, deadlocków i streamingu.
