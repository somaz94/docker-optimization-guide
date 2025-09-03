# main.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import pandas as pd
import numpy as np
from datetime import datetime
import uvicorn
import os

app = FastAPI(title="ML Prediction API", version="1.0.0")

# 데이터 모델
class PredictionRequest(BaseModel):
    features: list[float]
    model_name: str = "linear"

class PredictionResponse(BaseModel):
    prediction: float
    confidence: float
    timestamp: str

# 간단한 ML 모델 시뮬레이션
class SimpleModel:
    def __init__(self):
        # 실제로는 pickle로 로드된 모델일 것
        self.weights = np.random.randn(10)
        
    def predict(self, features):
        if len(features) != len(self.weights):
            raise ValueError("Feature dimension mismatch")
        
        prediction = np.dot(features, self.weights)
        confidence = min(0.95, abs(prediction) / 10)
        return prediction, confidence

model = SimpleModel()

@app.get("/")
async def root():
    return {
        "service": "ML Prediction API",
        "status": "healthy",
        "timestamp": datetime.now().isoformat()
    }

@app.get("/health")
async def health_check():
    return {
        "status": "ok",
        "memory_usage": pd.DataFrame({"test": [1, 2, 3]}).memory_usage(deep=True).sum()
    }

@app.post("/predict", response_model=PredictionResponse)
async def predict(request: PredictionRequest):
    try:
        if len(request.features) != 10:
            raise HTTPException(status_code=400, detail="Expected 10 features")
            
        prediction, confidence = model.predict(request.features)
        
        return PredictionResponse(
            prediction=float(prediction),
            confidence=float(confidence),
            timestamp=datetime.now().isoformat()
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    port = int(os.getenv("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
