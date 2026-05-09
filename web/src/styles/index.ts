// 样式入口文件
import './global.scss'

// 导入 Google Fonts (在 index.html 中也可以通过 link 标签引入)
export const loadFonts = () => {
  const link = document.createElement('link')
  link.rel = 'stylesheet'
  link.href = 'https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap'
  document.head.appendChild(link)
}
