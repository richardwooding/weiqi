# weiqi

Go / Weiqi rules engine — pure logic, no protocol. Played 9×9 with area
(Chinese) scoring and komi 6.5 to white. (The package is named "weiqi"
because "go" is a reserved word.)

Covers stone placement with capture, suicide and positional-ko rejection,
legal-move enumeration, flood-fill territory counting, and final area
scoring. Boards are flat `[81]int8` values (0 empty, 1 black, 2 white), so
positions copy cheaply and everything stays allocation-light.

Known limitation: no dead-stone negotiation — two consecutive passes score
the board exactly as it stands (area scoring makes filling your own
territory free, so capturing or filling dead stones first is always safe).

```go
b, captured, err := weiqi.ApplyMove(board, colour, idx, forbidden) // forbidden = ko board
moves := weiqi.LegalMoves(board, colour, forbidden)
black, white, winner := weiqi.FinalScore(board)
```

Deterministic, dependency-free, compiles to WASM. Extracted from
[kibitz](https://github.com/richardwooding/kibitz).

MIT licensed.
