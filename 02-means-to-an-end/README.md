# Test

```
go run main.go
xxd -r -p test.hex | nc localhost 9000 | hexdump
```
