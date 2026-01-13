"""
Pydantic models for ML sidecar API

Defines request/response schemas for all ML endpoints.
"""

from typing import List, Dict, Any, Optional
from pydantic import BaseModel, Field, field_validator, ConfigDict


class HealthResponse(BaseModel):
    """Health check response"""
    model_config = ConfigDict(protected_namespaces=())
    status: str
    version: str
    model_loaded: bool


class TrainRequest(BaseModel):
    """Request to train a model"""
    features: List[Dict[str, Any]] = Field(
        ...,
        description="List of feature dictionaries for each training sample"
    )
    labels: List[int] = Field(
        ...,
        description="List of profile IDs (labels) for each sample"
    )
    model_type: str = Field(
        ...,
        description="Model type: PROFILE_CLASSIFIER or SESSION_CLUSTERER"
    )

    @field_validator('model_type')
    @classmethod
    def validate_model_type(cls, v: str) -> str:
        if v not in ['PROFILE_CLASSIFIER', 'SESSION_CLUSTERER']:
            raise ValueError(f"Invalid model_type: {v}")
        return v

    @field_validator('features', 'labels')
    @classmethod
    def validate_lengths(cls, v: List, info) -> List:
        if not v:
            raise ValueError("Cannot be empty")
        return v


class TrainResponse(BaseModel):
    """Response from training"""
    model_config = ConfigDict(protected_namespaces=())
    success: bool
    model_version: str
    algorithm: str
    metrics: Dict[str, float] = Field(
        default_factory=dict,
        description="Training metrics (accuracy, precision, recall, etc.)"
    )
    samples_trained: int
    message: Optional[str] = None


class PredictRequest(BaseModel):
    """Request to predict profile assignments"""
    features: List[Dict[str, Any]] = Field(
        ...,
        description="List of feature dictionaries for each block to predict"
    )
    threshold: float = Field(
        default=0.6,
        ge=0.0,
        le=1.0,
        description="Minimum confidence threshold (0.0-1.0)"
    )


class PredictionResult(BaseModel):
    """Single prediction result"""
    block_index: int = Field(
        ...,
        description="Index in the input features list"
    )
    predicted_profile_id: int
    confidence: float = Field(
        ...,
        ge=0.0,
        le=1.0
    )
    confidence_level: str = Field(
        ...,
        description="HIGH, MEDIUM, or LOW"
    )


class PredictResponse(BaseModel):
    """Response from prediction"""
    model_config = ConfigDict(protected_namespaces=())
    success: bool
    predictions: List[PredictionResult]
    model_version: str
    total_predictions: int
    message: Optional[str] = None


class BlockData(BaseModel):
    """Block data for clustering"""
    block_id: int
    ts_start: str = Field(
        ...,
        description="ISO-8601 timestamp"
    )
    ts_end: str = Field(
        ...,
        description="ISO-8601 timestamp"
    )
    app_name: Optional[str] = None
    title: Optional[str] = None


class ClusterRequest(BaseModel):
    """Request to cluster blocks into sessions"""
    blocks: List[BlockData] = Field(
        ...,
        description="List of blocks with timestamps"
    )
    gap_threshold_minutes: int = Field(
        default=30,
        ge=1,
        le=480,
        description="Maximum gap within a session (1-480 minutes)"
    )

    @field_validator('blocks')
    @classmethod
    def validate_blocks(cls, v: List[BlockData]) -> List[BlockData]:
        if not v:
            raise ValueError("Cannot cluster empty block list")
        return v


class SessionData(BaseModel):
    """Session cluster result"""
    session_id: int
    block_ids: List[int]
    start_time: str
    end_time: str
    duration_minutes: float
    block_count: int


class ClusterResponse(BaseModel):
    """Response from clustering"""
    success: bool
    sessions: List[SessionData]
    total_blocks: int
    total_sessions: int
    message: Optional[str] = None
