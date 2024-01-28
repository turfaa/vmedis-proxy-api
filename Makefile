protoc:
	protoc -I=kafkapb --go_out=paths=source_relative:kafkapb kafkapb/*.proto