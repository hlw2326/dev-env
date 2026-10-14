# 软件 Logo 资源目录 (Software Logos Directory)

请将各软件的品牌 Logo 图片（推荐 SVG 或透明背景 PNG 格式）放置于此目录中。

### 推荐文件命名规范：
- Nginx: `nginx.svg` 或 `nginx.png`
- Apache: `apache.svg` 或 `apache.png`
- PHP: `php.svg` 或 `php.png`
- Node.js: `nodejs.svg` 或 `nodejs.png`
- MySQL: `mysql.svg` 或 `mysql.png`
- Redis: `redis.svg` 或 `redis.png`

### 说明：
当放置对应的 SVG/PNG 文件后，全局组件 `<h-logo :kind="kind" />` 将自动加载对应图标。如果未放置或图片不存在，系统将自动显示精致的品牌徽标缩写占位符，绝不显示破损图片。
