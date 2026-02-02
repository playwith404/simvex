# SIMVEX

공학 학습용 웹 기반 3D 기계 부품 뷰어.

## 구조
- `frontend/`: React + Vite + Three.js
- `backend/`: Go + Gin + PostgreSQL

## 로컬 실행

### 1) 백엔드
```bash
cd backend
export OPENAI_API_KEY=sk-...
export OPENAI_MODEL=gpt-4o-mini
export PORT=8080
export DATABASE_URL=postgres://simvex:simvex@localhost:5432/simvex?sslmode=disable
# 로컬 자산을 DB로 적재하려면
export ASSET_IMPORT_PATH=./assets/models
# go가 설치되어 있어야 합니다.
go run ./cmd/server
```

### 2) 프론트엔드
```bash
cd frontend
npm install
npm run dev
```

Vite 프록시가 `/api`와 `/assets`를 `http://localhost:8080`으로 전달합니다.

## 도커 실행
```bash
cp .env.example .env
# .env에 OPENAI_API_KEY 설정

docker-compose up --build
```

## 자산
3D 모델은 `/root/dosa/3D Asset`에서 `/root/dosa/simvex/backend/assets/models`로 복사되어 있습니다.
앱 실행 시 DB에 자산이 없으면 `ASSET_IMPORT_PATH` 경로의 파일을 PostgreSQL `assets` 테이블로 적재합니다.

## API
- `GET /api/objects`
- `GET /api/objects/:id`
- `GET /api/objects/:id/parts`
- `GET /api/parts/:partId`
- `POST /api/ai/chat`
