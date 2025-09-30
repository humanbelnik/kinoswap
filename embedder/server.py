import grpc
from concurrent import futures
import logging
import proto.embedding_pb2 as embedding_pb2
import proto.embedding_pb2_grpc as embedding_pb2_grpc
from service import EmbeddingService

class EmbeddingServicer(embedding_pb2_grpc.EmbeddingServiceServicer):
    def __init__(self):
        self.embedding_service = EmbeddingService()
        self.logger = logging.getLogger(__name__)
        self.logger.info("EmbeddingServicer initialized")

    def CreateMovieEmbedding(self, request, context):
        try:
            genres_text = ", ".join(request.genres) if request.genres else "unknown"
            text_data = f"Title: {request.title}. Genres: {genres_text}. Overview: {request.overview}. Year: {request.year}. Rating: {request.rating}"
            
            self.logger.info(f"Creating movie embedding for: {request.title}")
        
            embedding = self.embedding_service.build_embedding(text_data)
            
            embedding_list = embedding.tolist()
            
            self.logger.info(f"Successfully created embedding with dimension: {len(embedding_list)}")
            
            return embedding_pb2.EmbeddingResponse(embedding=embedding_list) # pyright: ignore[reportAttributeAccessIssue]
            
        except Exception as e:
            self.logger.error(f"Error in CreateMovieEmbedding: {str(e)}")
            context.set_details(f"Internal server error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return embedding_pb2.EmbeddingResponse() # pyright: ignore[reportAttributeAccessIssue]

    def CreatePreferenceEmbedding(self, request, context):
        try:
            self.logger.info(f"Creating preference embedding for text: {request.text[:100]}...")
            
            embedding = self.embedding_service.build_embedding(request.text)
            embedding_list = embedding.tolist()
            
            self.logger.info(f"Successfully created embedding with dimension: {len(embedding_list)}")
            
            return embedding_pb2.EmbeddingResponse(embedding=embedding_list) # pyright: ignore[reportAttributeAccessIssue]
            
        except Exception as e:
            self.logger.error(f"Error in CreatePreferenceEmbedding: {str(e)}")
            context.set_details(f"Internal server error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return embedding_pb2.EmbeddingResponse() # pyright: ignore[reportAttributeAccessIssue]

def serve():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        options=[
            ('grpc.max_send_message_length', 50 * 1024 * 1024),
            ('grpc.max_receive_message_length', 50 * 1024 * 1024),
        ]
    )
    
    embedding_pb2_grpc.add_EmbeddingServiceServicer_to_server(
        EmbeddingServicer(), server
    )
    
    port = '50051'
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    
    logging.info(f"gRPC Server started on port {port}")
    logging.info("Press Ctrl+C to stop the server")
    
    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logging.info("Shutting down server...")
        server.stop(0)

if __name__ == '__main__':
    serve()