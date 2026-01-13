"""
ML Service - Core machine learning logic

Implements:
- TF-IDF feature extraction from window titles
- Logistic Regression classifier with probability calibration
- Temporal session clustering
- Confidence scoring (HIGH/MEDIUM/LOW)

This service is compute-only - no database access.
"""

import logging
from typing import List, Dict, Any, Tuple
from datetime import datetime
import numpy as np
import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.calibration import CalibratedClassifierCV
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, precision_recall_fscore_support

from .models import (
    TrainResponse,
    PredictResponse,
    PredictionResult,
    ClusterResponse,
    SessionData,
    BlockData
)

logger = logging.getLogger(__name__)


class MLService:
    """
    ML Service for ChronicleCore

    Provides training, prediction, and clustering capabilities.
    Stateless - all state is provided via API calls.
    """

    def __init__(self):
        """Initialize ML service"""
        self.vectorizer = None
        self.classifier = None
        self.model_version = None
        self.model_loaded = False

        logger.info("MLService initialized")

    def train(
        self,
        features: List[Dict[str, Any]],
        labels: List[int],
        model_type: str
    ) -> TrainResponse:
        """
        Train a classification model

        Args:
            features: List of feature dicts with 'title', 'app_name', etc.
            labels: List of profile IDs (ground truth)
            model_type: 'PROFILE_CLASSIFIER' or 'SESSION_CLUSTERER'

        Returns:
            TrainResponse with metrics and model info
        """
        if len(features) != len(labels):
            raise ValueError("Features and labels length mismatch")

        if len(features) < 10:
            raise ValueError("Need at least 10 samples to train")

        logger.info(f"Training {model_type} with {len(features)} samples")

        # Extract text features
        texts = self._extract_text_features(features)

        # Split for validation
        # Note: stratify=labels removed to prevent errors when some classes have too few samples (<2)
        try:
            X_train, X_val, y_train, y_val = train_test_split(
                texts, labels, test_size=0.2, random_state=42
            )
        except Exception as e:
            logger.error(f"Split failed: {e}")
            # Fallback to simple slice if split fails
            split_idx = int(len(texts) * 0.8)
            X_train, X_val = texts[:split_idx], texts[split_idx:]
            y_train, y_val = labels[:split_idx], labels[split_idx:]

        # TF-IDF vectorization
        self.vectorizer = TfidfVectorizer(
            max_features=500,
            ngram_range=(1, 2),
            min_df=2,
            lowercase=True,
            strip_accents='unicode'
        )

        X_train_vec = self.vectorizer.fit_transform(X_train)
        X_val_vec = self.vectorizer.transform(X_val)

        # Train Logistic Regression
        base_classifier = LogisticRegression(
            max_iter=1000,
            multi_class='multinomial',
            random_state=42,
            class_weight='balanced'
        )

        # Calibrate probabilities for better confidence estimates
        self.classifier = CalibratedClassifierCV(
            base_classifier,
            method='sigmoid',
            cv=3
        )

        self.classifier.fit(X_train_vec, y_train)

        # Evaluate
        y_pred = self.classifier.predict(X_val_vec)
        accuracy = accuracy_score(y_val, y_pred)

        precision, recall, f1, _ = precision_recall_fscore_support(
            y_val, y_pred, average='weighted', zero_division=0
        )

        # Generate model version
        self.model_version = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
        self.model_loaded = True

        metrics = {
            "accuracy": float(accuracy),
            "precision": float(precision),
            "recall": float(recall),
            "f1_score": float(f1),
            "train_samples": len(X_train),
            "val_samples": len(X_val)
        }

        logger.info(f"Training complete: accuracy={accuracy:.3f}, f1={f1:.3f}")

        return TrainResponse(
            success=True,
            model_version=self.model_version,
            algorithm="TF-IDF + Calibrated Logistic Regression",
            metrics=metrics,
            samples_trained=len(features),
            message=f"Model trained successfully with {accuracy:.2%} accuracy"
        )

    def predict(
        self,
        features: List[Dict[str, Any]],
        threshold: float = 0.6
    ) -> PredictResponse:
        """
        Predict profile assignments for blocks

        Args:
            features: List of feature dicts to predict
            threshold: Minimum confidence threshold (0.0-1.0)

        Returns:
            PredictResponse with predictions and confidence scores
        """
        if not self.model_loaded:
            raise ValueError("No model loaded - train first")

        if not features:
            raise ValueError("No features provided for prediction")

        logger.info(f"Predicting for {len(features)} blocks (threshold={threshold})")

        # Extract text features
        texts = self._extract_text_features(features)

        # Vectorize
        X_vec = self.vectorizer.transform(texts)

        # Predict with probabilities
        predictions = self.classifier.predict(X_vec)
        probabilities = self.classifier.predict_proba(X_vec)

        # Build results
        results = []
        for i, (pred_profile_id, probs) in enumerate(zip(predictions, probabilities)):
            # Max probability for predicted class
            confidence = float(np.max(probs))

            # Only include predictions above threshold
            if confidence >= threshold:
                confidence_level = self._confidence_level(confidence)

                results.append(PredictionResult(
                    block_index=i,
                    predicted_profile_id=int(pred_profile_id),
                    confidence=confidence,
                    confidence_level=confidence_level
                ))

        logger.info(f"Generated {len(results)} predictions above threshold")

        return PredictResponse(
            success=True,
            predictions=results,
            model_version=self.model_version,
            total_predictions=len(results),
            message=f"Predicted {len(results)}/{len(features)} blocks above threshold"
        )

    def cluster(
        self,
        blocks: List[BlockData],
        gap_threshold_minutes: int = 30
    ) -> ClusterResponse:
        """
        Cluster blocks into sessions based on temporal gaps

        Args:
            blocks: List of blocks with timestamps
            gap_threshold_minutes: Maximum gap within a session

        Returns:
            ClusterResponse with session groupings
        """
        if not blocks:
            raise ValueError("No blocks provided for clustering")

        logger.info(f"Clustering {len(blocks)} blocks (gap={gap_threshold_minutes}m)")

        # Convert to DataFrame for easier manipulation
        df = pd.DataFrame([
            {
                'block_id': b.block_id,
                'ts_start': pd.to_datetime(b.ts_start),
                'ts_end': pd.to_datetime(b.ts_end)
            }
            for b in blocks
        ])

        # Sort by start time
        df = df.sort_values('ts_start').reset_index(drop=True)

        # Calculate gaps between consecutive blocks
        df['gap_minutes'] = (df['ts_start'] - df['ts_end'].shift(1)).dt.total_seconds() / 60.0

        # Mark session breaks (gap > threshold or first block)
        df['session_break'] = (df['gap_minutes'] > gap_threshold_minutes) | (df['gap_minutes'].isna())

        # Assign session IDs
        df['session_id'] = df['session_break'].cumsum()

        # Group by session
        sessions = []
        for session_id, group in df.groupby('session_id'):
            block_ids = group['block_id'].tolist()
            start_time = group['ts_start'].min()
            end_time = group['ts_end'].max()
            duration = (end_time - start_time).total_seconds() / 60.0

            sessions.append(SessionData(
                session_id=int(session_id),
                block_ids=block_ids,
                start_time=start_time.isoformat() + 'Z',
                end_time=end_time.isoformat() + 'Z',
                duration_minutes=float(duration),
                block_count=len(block_ids)
            ))

        logger.info(f"Clustered into {len(sessions)} sessions")

        return ClusterResponse(
            success=True,
            sessions=sessions,
            total_blocks=len(blocks),
            total_sessions=len(sessions),
            message=f"Clustered {len(blocks)} blocks into {len(sessions)} sessions"
        )

    def _extract_text_features(self, features: List[Dict[str, Any]]) -> List[str]:
        """
        Extract text from feature dicts for vectorization

        Args:
            features: List of feature dicts

        Returns:
            List of text strings
        """
        texts = []
        for f in features:
            parts = []

            # App name (important signal)
            if 'app_name' in f and f['app_name']:
                parts.append(f['app_name'])

            # Title (primary signal)
            if 'title' in f and f['title']:
                parts.append(f['title'])

            # Domain (if available)
            if 'domain' in f and f['domain']:
                parts.append(f['domain'])

            # Combine all text
            text = ' '.join(parts) if parts else 'unknown'
            texts.append(text)

        return texts

    @staticmethod
    def _confidence_level(confidence: float) -> str:
        """
        Convert confidence score to level

        Args:
            confidence: Probability 0.0-1.0

        Returns:
            'HIGH', 'MEDIUM', or 'LOW'
        """
        if confidence >= 0.85:
            return 'HIGH'
        elif confidence >= 0.60:
            return 'MEDIUM'
        else:
            return 'LOW'
