<template>
  <div class="landing-container">
    <!-- 电路背景动画 -->
    <div class="circuit-background">
      <svg class="circuit-svg" width="100%" height="100%" ref="circuitSvg"></svg>
    </div>

    <!-- 主要内容区域 -->
    <div class="content-wrapper">
      <header class="landing-header">
        <div class="logo">
          <h1>Virtual Human Studio</h1>
        </div>
        <div class="nav-buttons">
          <el-button type="text" @click="$router.push('/contact')">联系我们</el-button>
          <el-button type="text" @click="$router.push('/login')">登录</el-button>
          <el-button type="primary" @click="$router.push('/register')">注册</el-button>
        </div>
      </header>

      <main class="main-content">
        <section class="hero-section">
          <h2>下一代数字人创作平台</h2>
          <p>打造您的专属数字人，让AI为您的创意赋能</p>
          <el-button type="primary" size="large" @click="$router.push('/register')">立即开始</el-button>
        </section>

        <section class="features-section">
          <h3>核心功能</h3>
          <div class="feature-grid">
            <div class="feature-card" v-for="(feature, index) in features" :key="index">
              <i :class="feature.icon"></i>
              <h4>{{ feature.title }}</h4>
              <p>{{ feature.description }}</p>
            </div>
          </div>
        </section>

        <section class="tech-section">
          <h3>技术优势</h3>
          <div class="tech-grid">
            <div class="tech-item" v-for="(tech, index) in techAdvantages" :key="index">
              <div class="tech-icon">
                <i :class="tech.icon"></i>
              </div>
              <div class="tech-content">
                <h4>{{ tech.title }}</h4>
                <p>{{ tech.description }}</p>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Landing',
  data() {
    return {
      features: [
        {
          icon: 'el-icon-microphone',
          title: '音色克隆',
          description: '通过先进的AI技术，完美复制您的声音特征'
        },
        {
          icon: 'el-icon-reading',
          title: '文本转语音',
          description: '将文字转换为自然流畅的语音输出'
        },
        {
          icon: 'el-icon-video-camera',
          title: '数字人合成',
          description: '生成高度逼真的数字人视频内容'
        },
        {
          icon: 'el-icon-collection',
          title: '资源管理',
          description: '便捷管理您的音色和数字人资源'
        }
      ],
      techAdvantages: [
        {
          icon: 'el-icon-cpu',
          title: '先进算法',
          description: '采用最新的深度学习技术，确保生成内容的高质量和真实性'
        },
        {
          icon: 'el-icon-lightning',
          title: '实时处理',
          description: '高效的处理引擎，支持快速生成和实时预览'
        },
        {
          icon: 'el-icon-magic-stick',
          title: '智能定制',
          description: '个性化参数调整，打造独特的数字人形象'
        }
      ]
    }
  },
  mounted() {
    this.initCircuitAnimation()
    window.addEventListener('mousemove', this.handleMouseMove)
  },
  beforeDestroy() {
    window.removeEventListener('mousemove', this.handleMouseMove)
  },
  methods: {
    initCircuitAnimation() {
      const svg = this.$refs.circuitSvg
      const width = svg.clientWidth
      const height = svg.clientHeight
      
      // 创建电路节点
      const nodes = []
      for (let i = 0; i < 20; i++) {
        nodes.push({
          x: Math.random() * width,
          y: Math.random() * height
        })
      }
      
      // 绘制电路线
      nodes.forEach((node, index) => {
        if (index < nodes.length - 1) {
          const line = document.createElementNS('http://www.w3.org/2000/svg', 'path')
          const x1 = node.x
          const y1 = node.y
          const x2 = nodes[index + 1].x
          const y2 = nodes[index + 1].y
          
          // 创建折线路径
          const midX = (x1 + x2) / 2
          const path = `M ${x1} ${y1} L ${midX} ${y1} L ${midX} ${y2} L ${x2} ${y2}`
          
          line.setAttribute('d', path)
          line.setAttribute('stroke', 'rgba(0, 255, 255, 0.2)')
          line.setAttribute('stroke-width', '2')
          line.setAttribute('fill', 'none')
          
          // 添加流动动画
          const animation = document.createElementNS('http://www.w3.org/2000/svg', 'animate')
          animation.setAttribute('attributeName', 'stroke-dashoffset')
          animation.setAttribute('from', '1000')
          animation.setAttribute('to', '0')
          animation.setAttribute('dur', '3s')
          animation.setAttribute('repeatCount', 'indefinite')
          
          line.style.strokeDasharray = '10, 10'
          line.appendChild(animation)
          svg.appendChild(line)
        }
      })
    },
    handleMouseMove(event) {
      const mouseX = event.clientX
      const mouseY = event.clientY
      
      // 获取所有电路线
      const lines = this.$refs.circuitSvg.querySelectorAll('path')
      
      lines.forEach(line => {
        const rect = line.getBoundingClientRect()
        const distance = Math.sqrt(
          Math.pow(mouseX - (rect.left + rect.width / 2), 2) +
          Math.pow(mouseY - (rect.top + rect.height / 2), 2)
        )
        
        // 根据鼠标距离调整线条亮度
        const opacity = Math.max(0.2, 1 - distance / 500)
        line.style.stroke = `rgba(0, 255, 255, ${opacity})`
      })
    }
  }
}
</script>

<style scoped>
.landing-container {
  min-height: 100vh;
  position: relative;
  background: linear-gradient(135deg, #1a1a1a, #000);
  color: #fff;
  overflow: hidden;
}

.circuit-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
  opacity: 0.5;
}

.content-wrapper {
  position: relative;
  z-index: 2;
  padding: 2rem;
}

.landing-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 0;
}

.logo h1 {
  font-size: 2rem;
  background: linear-gradient(120deg, #00ffff, #0099ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0;
}

.nav-buttons .el-button {
  margin-left: 1rem;
  font-size: 1.1rem;
}

.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 4rem 0;
}

.hero-section {
  text-align: center;
  margin-bottom: 6rem;
}

.hero-section h2 {
  font-size: 3.5rem;
  margin-bottom: 1.5rem;
  background: linear-gradient(120deg, #00ffff, #0099ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-section p {
  font-size: 1.5rem;
  color: #a0a0a0;
  margin-bottom: 2rem;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
  margin-top: 3rem;
}

.feature-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 15px;
  padding: 2rem;
  text-align: center;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-10px);
  background: rgba(255, 255, 255, 0.1);
}

.feature-card i {
  font-size: 3rem;
  color: #00ffff;
  margin-bottom: 1.5rem;
}

.feature-card h4 {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: #fff;
}

.feature-card p {
  color: #a0a0a0;
  line-height: 1.6;
}

.tech-section {
  margin-top: 6rem;
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 3rem;
  margin-top: 3rem;
}

.tech-item {
  display: flex;
  align-items: flex-start;
  gap: 1.5rem;
}

.tech-icon {
  font-size: 2.5rem;
  color: #00ffff;
}

.tech-content h4 {
  font-size: 1.3rem;
  margin-bottom: 0.5rem;
  color: #fff;
}

.tech-content p {
  color: #a0a0a0;
  line-height: 1.6;
}

h3 {
  font-size: 2.5rem;
  text-align: center;
  margin-bottom: 2rem;
  background: linear-gradient(120deg, #00ffff, #0099ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

@media (max-width: 768px) {
  .hero-section h2 {
    font-size: 2.5rem;
  }
  
  .hero-section p {
    font-size: 1.2rem;
  }
  
  .feature-grid {
    grid-template-columns: 1fr;
  }
  
  .tech-grid {
    grid-template-columns: 1fr;
  }
}
</style>