## 🔧 Push'n'Pray

> Light a candle for your deploy. Bless your CI with hope and memes 🙏

![Example](https://yourdomain.com/candles/example-id/image)

[![Docker Image CI](https://github.com/your-org/your-repo/actions/workflows/docker.yml/badge.svg)](https://github.com/your-org/your-repo/actions/workflows/docker.yml)
[![Demo](https://img.shields.io/badge/demo-live-green)](https://yourdomain.com)

---

### 📆 What is this?

**Push'n'Pray** is a fun service that lets you "light a candle" before or after a release. It generates a **blessed animated candle GIF** with your message and a unique link you can share or embed in your Merge Request, Slack, or README.

---

### ⚙️ Features

- 🕯️ Animated candle with flickering flame
- ✍️ Custom message and author
- 🌐 Shareable image URL
- 📎 Markdown embed code for Merge Requests
- 🟣 Docker-ready + CI/CD friendly

---

### 🚀 How to use

1. Visit the web UI (or call API)
2. Enter your name and a wish (e.g. _“Please no 500 on prod 🙏”_)
3. Click "Submit candle"
4. Embed the result:

```md
![Blessed Candle](https://push-n-pray.tech-pioneer.info)
```

---

### 🟣 Deployment

#### Docker

```bash
docker build -t push-n-pray .
docker run -p 8080:8080 push-n-pray
```

Visit [http://localhost:8080](http://localhost:8080)

---

### 💥 Environment Variables

| Name       | Description            | Default |
|------------|------------------------|---------|
| `PORT`     | Port to run the server | `8080`  |

---

### 🧪 API

#### `POST /candles`

Create a new candle.

**Form fields:**

- `author` — your name
- `message` — your wish or blessing

**Response (HTML):** returns a preview with markdown embed

---

### 📅 GitHub Actions

This project includes a GitHub Action to build and publish Docker images to `ghcr.io`.

See `.github/workflows/docker.yml`

---

### ❤️ Credits

Made with memes, Go and love by your friendly DevOps priest.  
Bless your pipelines. Amen. 🙏

