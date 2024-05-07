package services

import (
	"context"
	"flag"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr                  = flag.String("addr", "localhost:6334", "the address to connect to")
	collectionName        = "test_collection"
	vectorSize     uint64 = 4
	distance              = pb.Distance_Dot
)

type QdrantService interface {
	ListCollections() ([]string, error)
	CreateCollection(collectionName string, vectorSize uint64, distance pb.Distance, defaultSegmentNumber uint64) error
}

type qdrantServiceImpl struct {
	collectionsClient pb.CollectionsClient
}

func NewQdrantService(addr string) (QdrantService, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	collectionsClient := pb.NewCollectionsClient(conn)

	return &qdrantServiceImpl{
		collectionsClient: collectionsClient,
	}, nil
}

func (s *qdrantServiceImpl) ListCollections() ([]string, error) {
    ctx := context.Background()
    r, err := s.collectionsClient.List(ctx, &pb.ListCollectionsRequest{})
    if err != nil {
        return nil, err
    }

    collections := make([]string, len(r.GetCollections()))

    for i, collection := range r.GetCollections() {
        collections[i] = collection.GetName()
    }

    return collections, nil
}

func (s *qdrantServiceImpl) CreateCollection(collectionName string, vectorSize uint64, distance pb.Distance, defaultSegmentNumber uint64) error {
	ctx := context.Background()
	_, err := s.collectionsClient.Create(ctx, &pb.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &pb.VectorsConfig{Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     vectorSize,
				Distance: distance,
			},
		}},
		OptimizersConfig: &pb.OptimizersConfigDiff{
			DefaultSegmentNumber: &defaultSegmentNumber,
		},
	})
	return err
}

