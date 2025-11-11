import logging
from sentence_transformers import SentenceTransformer
import numpy as np

class EmbeddingService:
    def __init__(self, model_name: str = "./model/"):

        

        self.logger = logging.getLogger(__name__)
        self.logger.info(f"Loading model: {model_name}")
        self.model = SentenceTransformer(model_name)
        self.logger.info("Model loaded successfully")

    def build_embedding(self, text: str) -> np.ndarray:
        try:
            embedding = self.model.encode(text)
            return embedding # pyright: ignore[reportReturnType]
        except Exception as e:
            self.logger.error(f"Error generating embedding: {str(e)}")
            raise