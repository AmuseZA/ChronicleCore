"""
Authentication module for ML sidecar

Validates X-CC-Token header against expected token.
Token should match CC_ML_TOKEN environment variable.
"""

import os
from fastapi import HTTPException

def verify_token(token: str) -> None:
    """
    Verify that the provided token matches the expected ML token.

    Args:
        token: Token from X-CC-Token header

    Raises:
        HTTPException: 401 if token is invalid or missing
    """
    expected_token = os.getenv("CC_ML_TOKEN")

    if not expected_token:
        raise HTTPException(
            status_code=500,
            detail="ML service not configured (CC_ML_TOKEN missing)"
        )

    if not token:
        raise HTTPException(
            status_code=401,
            detail="Missing X-CC-Token header"
        )

    if token != expected_token:
        raise HTTPException(
            status_code=401,
            detail="Invalid authentication token"
        )
