import grpc
from concurrent import futures
import logging
import embedder_pb2 as embedding_pb2
import embedder_pb2_grpc as embedding_pb2_grpc 
from service import EmbeddingService
import os
from flask import Flask, request, jsonify

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
            
            return embedding_pb2.EmbeddingResponse(embedding=embedding_list)
            
        except Exception as e:
            self.logger.error(f"Error in CreateMovieEmbedding: {str(e)}")
            context.set_details(f"Internal server error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return embedding_pb2.EmbeddingResponse()

    def CreatePreferenceEmbedding(self, request, context):
        try:
            self.logger.info(f"Creating preference embedding for text: {request.text[:100]}...")
            
            embedding = self.embedding_service.build_embedding(request.text)
            embedding_list = embedding.tolist()
            
            self.logger.info(f"Successfully created embedding with dimension: {len(embedding_list)}")
            
            return embedding_pb2.EmbeddingResponse(embedding=embedding_list)
            
        except Exception as e:
            self.logger.error(f"Error in CreatePreferenceEmbedding: {str(e)}")
            context.set_details(f"Internal server error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return embedding_pb2.EmbeddingResponse()

def create_flask_app():
    app = Flask(__name__)
    embedding_service = EmbeddingService()
    
    @app.route('/health', methods=['GET'])
    def health_check():
        return jsonify({"status": "healthy"})
    
    @app.route('/preference_embedding', methods=['POST'])
    def preference_embedding():
        try:
            data = request.get_json()
            if not data or 'text' not in data:
                return jsonify({"error": "Missing 'text' field in request body"}), 400
            
            text = data['text']
            logging.info(f"Creating preference embedding for text: {text[:100]}...")
            
            embedding = embedding_service.build_embedding(text)
            embedding_list = embedding.tolist()
            
            logging.info(f"Successfully created embedding with dimension: {len(embedding_list)}")
            
            return jsonify({"embedding": embedding_list})
            
        except Exception as e:
            logging.error(f"Error in preference_embedding: {str(e)}")
            return jsonify({"error": f"Internal server error: {str(e)}"}), 500
    
    return app

def serve_grpc():
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
    server.wait_for_termination()

def serve_http():
    app = create_flask_app()
    port = 5000
    logging.info(f"HTTP Server started on port {port}")
    app.run(host='0.0.0.0', port=port)

def serve(delivery_proto='GRPC'):
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    logging.info(f"proto={delivery_proto}")
    if delivery_proto.upper() == 'HTTP':
        serve_http()
    else:
        serve_grpc()