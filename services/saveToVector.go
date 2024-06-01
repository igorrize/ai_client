package services

import (
	"context"
	"flag"
	"github.com/google/uuid"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

var (
	waitUpsert     bool   = true
	addr                  = flag.String("addr", "localhost:6334", "the address to connect to")
	collectionName        = "test_collection"
	vectorSize     uint64 = 4
	distance              = pb.Distance_Dot
)

type QdrantService interface {
	ListCollections() ([]string, error)
	CreateCollection(collectionName string, vectorSize uint64, distance pb.Distance, defaultSegmentNumber uint64) error
	UpsertPoints(collectionName string, waitUpsert bool, upsertPoints [][]float32, payload []string) error
}

type qdrantServiceImpl struct {
	collectionsClient pb.CollectionsClient
	pointsClient      pb.PointsClient
}

func NewQdrantService(addr string) (QdrantService, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	collectionsClient := pb.NewCollectionsClient(conn)
	pointsClient := pb.NewPointsClient(conn)
	return &qdrantServiceImpl{
		collectionsClient: collectionsClient,
		pointsClient:      pointsClient,
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

func (s *qdrantServiceImpl) UpsertPoints(collectionName string, waitUpserts bool, points [][]float32, payload []string) error {
	upsertPoints := []*pb.PointStruct{}

	for i, payloadString := range payload {
		point := &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: uuid.New().String()},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: points[i]}}},
			Payload: map[string]*pb.Value{"text": {Kind: &pb.Value_StringValue{StringValue: payloadString}}}}
		upsertPoints = append(upsertPoints, point)
	}
	ctx := context.Background()
	_, err := s.pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})
	if err != nil {
		log.Fatalf("Could not upsert points: %v", err)
	} else {
		log.Println("Upsert", len(upsertPoints), "points")
	}

	return err
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
