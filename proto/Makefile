PROTO = sqlite_backup_server

FLAGS = --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./$(PROTO).proto


$(PROTO).pb.go: $(PROTO).proto
	protoc $(FLAGS)

