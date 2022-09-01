package main

import (
	"bufio"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/t4ke0/sqlite_backup_server/proto"
)

var (
	//go:embed tls.key
	tlsKeyFile string
	//go:embed tls.crt
	tlsCertFile string
	//
	serverPort string = os.Getenv("SERVER_PORT")
)

const mediaDir string = "media"

// BackupServer
type BackupServer struct {
	pb.UnimplementedBackupServiceServer
}

// Upload
func (b *BackupServer) Upload(stream pb.BackupService_UploadServer) error {
	var fd *os.File
	var touched bool
	for {
		req, err := stream.Recv()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !touched {
			filename := filepath.Join(mediaDir, req.FileName)
			var err error
			fd, err = os.Create(filename)
			if err != nil {
				return err
			}
			defer fd.Close()
			touched = true
		}
		if _, err := fd.Write(req.Chunk); err != nil {
			return err
		}
	}

	if err := stream.SendAndClose(&pb.UploadResponse{
		Ok: true,
	}); err != nil {
		return err
	}

	return nil
}

// Download
func (b *BackupServer) Download(req *pb.DownloadRequest, stream pb.BackupService_DownloadServer) error {
	filename := filepath.Join(mediaDir, req.FileName)
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	const chunkSize = 1024

	chunk := make([]byte, chunkSize)

	reader := bufio.NewReader(fd)
	for {
		n, err := reader.Read(chunk)
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.DownloadResponse{
			Chunk: chunk[:n],
		}); err != nil {
			return err
		}
	}

	return nil
}

func main() {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		log.Fatalf("Error failed to listen on %v [%v]", serverPort, err)
	}

	cert, err := tls.X509KeyPair([]byte(tlsCertFile), []byte(tlsKeyFile))
	if err != nil {
		log.Fatal(err)
	}

	creds := credentials.NewServerTLSFromCert(&cert)

	opts := []grpc.ServerOption{grpc.Creds(creds)}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterBackupServiceServer(grpcServer, &BackupServer{})

	log.Printf("listening on localhost:%v", serverPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpc server failed to initial serve %v", err)
	}
}
