docker-compose run --rm wait-benchmark-deps
go test -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -bench="BenchmarkPostgresSink" -benchmem -benchtime=10x -count=1 -timeout=3h -showModePrefix=true >  result/bench-all.out
go test -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -bench="BenchmarkPostgresSink" -benchmem -benchtime=10x -count=10 -timeout=3h -modes=default > result/bench-default.out
go test -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -bench="BenchmarkPostgresSink" -benchmem -benchtime=10x -count=10 -timeout=3h -modes=pipeline > result/bench-pipeline.out
go test -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -bench="BenchmarkPostgresSink" -benchmem -benchtime=10x -count=10 -timeout=3h -modes=native > result/bench-native.out

benchstat result/bench-default.out result/bench-pipeline.out result/bench-native.out > result/benchstat.out