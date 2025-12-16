# Eksperymenty (narzędzia)

Wszystkie komendy uruchamiaj z katalogu `backend/`.

## 1) Failover lidera (Bully)

- Generuje `experiments/bully.csv`
- Mierzy czas do odzyskania dostępności mastera z perspektywy klienta (`Ping()` z autodiscovery).

```bash
go run ./cmd/exp-bully -runs 10 -out experiments/bully.csv
```

## 2) Opóźnienie wejścia do sekcji krytycznej (lock)

- Generuje `experiments/lock.csv`
- W każdym run: wszystkie węzły proszą o lock na ten sam zasób, a harness zapisuje czas oczekiwania do wejścia.

```bash
go run ./cmd/exp-lock -runs 10 -nodes 5 -out experiments/lock.csv
```

## 3) Deadlock: czas rozstrzygnięcia

- Generuje `experiments/deadlock.csv`
- Tworzy klasyczny cykl: A trzyma `r1`, B trzyma `r2`, potem A prosi o `r2` i B o `r1`.
- Czas zawiera w sobie opóźnienie detekcji (monitor w masterze tyka co 10s).

```bash
go run ./cmd/exp-deadlock -runs 10 -out experiments/deadlock.csv
```

## 4) Streaming: upload/download throughput

- Generuje `experiments/streaming.csv`
- Upload oraz download pliku o zadanym rozmiarze.

```bash
go run ./cmd/exp-streaming -runs 5 -size-mb 64 -out experiments/streaming.csv
```
