# FEC Analysis

Analyzes Forward Error Correction (FEC) recovery probabilities under different network loss conditions using graph-based algorithms.

## What it does

Models packet loss scenarios as a state space graph where vertices represent sets of delivered/recovered packets and edges represent FEC-based recovery operations. Uses BFS to determine which loss patterns can be recovered for different FEC configurations.

## Requirements

- Go 1.24+

## Usage

```bash
# Run main analysis
go run cmd/fec-analysis/main.go

# Run tests
go test -v
```

## Project Structure

```
fec/
├── cmd/
│   ├── fec-analysis/       # Main analysis program
│   ├── loss-models-printer/
│   ├── matrix-printer/
│   └── graph-printer/
├── graph.go                # Graph interface and BFS
├── mask.go                 # FEC mask interface
├── recovery_graph.go       # Recovery graph implementation
├── googlebursty.go         # Google Bursty mask
├── googlerandom.go         # Google Random mask
├── loss_model.go           # Loss model interface
├── random_loss_model.go    # Random loss model
├── gilbert_elliott_loss_model.go  # Gilbert-Elliott model
└── recovery_characteristics.go    # Recovery metrics
```

## Components

### Masks
FEC protection patterns that determine which media packets are protected by each FEC packet.

- `googlebursty.go`: Optimized for burst losses
- `googlerandom.go`: Optimized for random losses

### Loss Models
- `RandomLossModel`: Independent packet loss with uniform probability
- `GilbertElliotLossModel`: 2-state Markov chain (good/bad states)

### Recovery Graph
Graph with 2^(N+K) vertices where:
- Each vertex is a bitset of delivered/recovered packets
- Bits 0 to N-1: Media packets
- Bits N to N+K-1: FEC packets
- Edges represent recovery operations

### Algorithm

1. Build recovery graph from FEC mask
2. Run multi-source BFS from "good" vertices (all N media packets present)
3. Calculate probability of each reachable state using loss model
4. Sum probabilities to get recovery probability

## Performance Notes

- State space grows as 2^(N+K), limiting analysis to N≤12
- BFS results are pre-computed and cached
- Gilbert-Elliott model uses dynamic programming with memoization
