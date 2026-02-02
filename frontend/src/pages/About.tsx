export const About = () => {
  return (
    <section className="about-page">
      <h2>SIMVEX 소개</h2>
      <p>
        SIMVEX는 공학 학습자를 위한 3D 기계 부품 뷰어입니다. 분해/조립, AI 질문,
        학습 노트 기능을 통해 기계 구조를 직관적으로 이해할 수 있습니다.
      </p>
      <div className="about-grid">
        <div>
          <h4>핵심 기능</h4>
          <ul>
            <li>3D 오브젝트 실시간 렌더링</li>
            <li>부품 선택 및 정보 표시</li>
            <li>AI 학습 어시스턴트</li>
            <li>워크플로우 차트</li>
          </ul>
        </div>
        <div>
          <h4>기술 스택</h4>
          <ul>
            <li>React + Vite + Three.js</li>
            <li>Go + Gin + SQLite</li>
            <li>OpenAI API 연동</li>
          </ul>
        </div>
      </div>
    </section>
  )
}
