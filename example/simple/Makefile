SERVICES=\
	another \
	some

$(SERVICES):
	protoc -I ./proto/$@/v1 -I . \
		--go_out=./proto/$@/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_out=./proto/$@/v1 \
		--go-grpc_opt=paths=source_relative \
		--openapiv2_out=logtostderr=true:./proto/$@ \
        --openapiv2_opt=output_format=yaml \
		--grpc-gateway_out=./proto/$@/v1 \
        --grpc-gateway_opt=logtostderr=true \
        --grpc-gateway_opt=paths=source_relative \
        --doc_out=./proto/$@/ \
        --doc_opt=markdown,$@.md \
        --experimental_allow_proto3_optional \
		$@.proto;