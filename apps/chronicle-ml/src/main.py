"""
ChronicleCore ML Sidecar

FastAPI service for ML predictions (compute-only, no direct DB access).
Authenticated via X-CC-Token header.
"""

import os
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Header, Depends
from fastapi.middleware.cors import CORSMiddleware

from .auth import verify_token
from .models import (
    TrainRequest, TrainResponse,
    PredictRequest, PredictResponse,
    ClusterRequest, ClusterResponse,
    HealthResponse
)
from .ml_service import MLService

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global ML service instance
ml_service = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifecycle manager for ML service"""
    global ml_service

    logger.info("ChronicleCore ML Sidecar starting...")

    # Initialize ML service
    ml_service = MLService()
    logger.info("ML service initialized")

    yield

    # Cleanup
    logger.info("ML service shutting down...")
    ml_service = None

# Create FastAPI app
app = FastAPI(
    title="ChronicleCore ML Sidecar",
    version="1.0.0",
    lifespan=lifespan
)

# CORS for localhost only
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://127.0.0.1", "http://localhost"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

def get_ml_service():
    """Dependency to get ML service"""
    if ml_service is None:
        raise HTTPException(status_code=503, detail="ML service not initialized")
    return ml_service

@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint (no auth required)"""
    return HealthResponse(
        status="ok",
        version="1.0.0",
        model_loaded=ml_service is not None and ml_service.model_loaded
    )

@app.post("/train", response_model=TrainResponse)
async def train_model(
    request: TrainRequest,
    token: str = Header(..., alias="X-CC-Token"),
    service: MLService = Depends(get_ml_service)
):
    """
    Train ML model on provided data

    Request body should contain:
    - features: List of feature dicts
    - labels: List of profile IDs
    - model_type: 'PROFILE_CLASSIFIER' or 'SESSION_CLUSTERER'
    """
    verify_token(token)

    try:
        logger.info(f"Training model with {len(request.features)} samples")

        result = service.train(
            features=request.features,
            labels=request.labels,
            model_type=request.model_type
        )

        logger.info(f"Training complete: accuracy={result.metrics.get('accuracy', 0):.3f}")

        return result

    except ValueError as e:
        logger.error(f"Training error: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"Training failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Training failed")

@app.post("/predict", response_model=PredictResponse)
async def predict(
    request: PredictRequest,
    token: str = Header(..., alias="X-CC-Token"),
    service: MLService = Depends(get_ml_service)
):
    """
    Predict profile assignments for given blocks

    Request body should contain:
    - features: List of feature dicts for each block
    - threshold: Minimum confidence (0.0-1.0)
    """
    verify_token(token)

    try:
        logger.info(f"Predicting for {len(request.features)} blocks")

        result = service.predict(
            features=request.features,
            threshold=request.threshold
        )

        logger.info(f"Predicted {len(result.predictions)} assignments")

        return result

    except ValueError as e:
        logger.error(f"Prediction error: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"Prediction failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Prediction failed")

@app.post("/cluster", response_model=ClusterResponse)
async def cluster_sessions(
    request: ClusterRequest,
    token: str = Header(..., alias="X-CC-Token"),
    service: MLService = Depends(get_ml_service)
):
    """
    Cluster blocks into sessions based on temporal gaps

    Request body should contain:
    - blocks: List of block dicts with timestamps
    - gap_threshold_minutes: Maximum gap within a session
    """
    verify_token(token)

    try:
        logger.info(f"Clustering {len(request.blocks)} blocks")

        result = service.cluster(
            blocks=request.blocks,
            gap_threshold_minutes=request.gap_threshold_minutes
        )

        logger.info(f"Found {len(result.sessions)} sessions")

        return result

    except ValueError as e:
        logger.error(f"Clustering error: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"Clustering failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail="Clustering failed")

if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("ML_PORT", "8081"))

    uvicorn.run(
        "src.main:app",
        host="127.0.0.1",
        port=port,
        log_level="info"
    )
