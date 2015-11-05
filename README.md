# num
An over-engineered package for adding thousands separators to numbers that supports streaming.

###### For example:
```
BenchmarkNum-4       20    62281610  ns/op  1161200  B/op  1  allocs/op
BenchmarkStream-4    20    62111283  ns/op  6572     B/op  0  allocs/op
```
###### Becomes:
```
BenchmarkNum-4       20    61,777,611  ns/op  1,161,200  B/op  1  allocs/op
BenchmarkStream-4    20    62,340,974  ns/op  6,572      B/op  0  allocs/op
```
