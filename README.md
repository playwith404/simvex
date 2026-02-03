# SIMVEX

공학 학습용 웹 기반 3D 기계 부품 뷰어. 3D 분해 학습, 노트/AI 대화, PDF 리포트, 워크플로우 차트를 제공합니다.

## 구성
- **Frontend**: React + Vite + Three.js + React Flow + jsPDF
- **Backend**: Go + Gin
- **DB**: PostgreSQL (자산 파일 포함)
- **Infra**: Docker + Nginx(HTTPS)
- **CI/CD**: GitHub Actions (frontend/backend 분리)

## 주요 기능
- 3D 뷰어 (회전/줌/팬, 부품 선택/하이라이트)
- 분해도 슬라이더 (부품별 분해 방향/거리)
- 노트/AI 채팅 탭
- PDF 리포트 (자동 카메라 피팅 + 한글 텍스트 이미지 렌더)
- 워크플로우 차트 (React Flow, 로컬 저장)

## 로컬 실행

### 1) DB 준비 (PostgreSQL)
```bash
createdb simvex
```

### 2) 백엔드
```bash
cd backend
export OPENAI_API_KEY=sk-...
export OPENAI_MODEL=gpt-4o-mini
export PORT=8080
export DATABASE_URL=postgres://simvex:simvex@localhost:5432/simvex?sslmode=disable
# 로컬 자산을 DB로 적재하려면
export ASSET_IMPORT_PATH=./assets/models

go run ./cmd/server
```

### 3) 프론트엔드
```bash
cd frontend
npm install
npm run dev
```

Vite에서 `/api`와 `/assets/models`는 `http://localhost:8080`으로 전달됩니다.

## 도커 실행
```bash
cp .env.example .env
# .env에 OPENAI_API_KEY, DATABASE_URL, OPENAI_MODEL 설정

docker compose up --build
```

기본 포트:
- Frontend: `http://localhost:8081`
- Backend: `http://localhost:8080`

## 자산 저장 방식
- 서버 시작 시 `ASSET_IMPORT_PATH` 경로를 스캔하여 `assets` 테이블에 적재합니다.
- 이후 `/assets/models/*` 경로로 DB에 저장된 바이너리를 제공합니다.

## 운영 (도메인/HTTPS)
- 도메인: `https://simvex.pjcloud.store`
- 호스트 Nginx가 TLS 종료
  - `/api/*` → `http://127.0.0.1:8080`
  - `/assets/models/*` → `http://127.0.0.1:8080`
  - `/` → `http://127.0.0.1:8081`

## CI/CD
- `.github/workflows/frontend-cicd.yml`
- `.github/workflows/backend-cicd.yml`

필수 Secrets:
- `SSH_HOST`, `SSH_USER`, `SSH_PRIVATE_KEY`
- `DATABASE_URL`, `OPENAI_MODEL`, `OPENAI_API_KEY`

## API
- `GET /api/objects`
- `GET /api/objects/:id`
- `GET /api/objects/:id/parts`
- `GET /api/parts/:partId`
- `POST /api/ai/chat`
- `GET /assets/models/*`

## 트러블슈팅
- PDF 저장 후 탭 크래시가 발생하면 HTTPS 접속 확인
- 3D 화면 크래시(Chrome 오류 코드 5)는 GPU/하드웨어 가속 이슈일 수 있음

---

© 2026 SIMVEX
