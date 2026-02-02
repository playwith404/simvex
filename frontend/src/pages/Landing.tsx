import { Link } from 'react-router-dom'

export const Landing = () => {
  return (
    <section className="landing">
      <div className="landing__hero">
        <div className="landing__hero-text">
          <p className="eyebrow">Engineering Learning Platform</p>
          <h1>공학 학습의 새로운 차원</h1>
          <p className="lead">
            3D 시각화로 기계 구조를 직관적으로 이해하고, AI와 함께 학습하세요.
          </p>
          <div className="landing__actions">
            <Link to="/objects" className="cta">학습 시작하기</Link>
            <Link to="/workflow" className="ghost">워크플로우 보기</Link>
          </div>
        </div>
        <div className="landing__hero-visual">
          <div className="hero-orb" />
        </div>
      </div>

      <div className="landing__features">
        <div className="feature">
          <h4>SIMVEX 3D View</h4>
          <p>분해/조립, 부품 선택, 실시간 학습 노트</p>
        </div>
        <div className="feature">
          <h4>3D 분해</h4>
          <p>부품별 분해 방향으로 구조를 단계적으로 학습합니다.</p>
        </div>
        <div className="feature">
          <h4>AI 질문</h4>
          <p>현재 부품 컨텍스트에 맞춘 답변으로 학습 효율을 높입니다.</p>
        </div>
        <div className="feature">
          <h4>노트 기록</h4>
          <p>학습 메모를 저장하고 PDF로 출력합니다.</p>
        </div>
      </div>
    </section>
  )
}
