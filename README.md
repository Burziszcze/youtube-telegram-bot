# 🎬 YouTube-Telegram Bot 📩

A **Go-based** Telegram bot that monitors YouTube channels and sends notifications about new videos to a specified Telegram channel or group.

---

## 📌 Features
✅ Monitor multiple YouTube channels  
✅ Send notifications to Telegram with a link to the new video  
✅ Automatically detect changes in `config.yml` – no need to restart the bot  
✅ Run in **Docker**  

---

## 📦 Installation & Setup

### 1️⃣ Clone the repository
```bash
git clone https://github.com/burziszcze/youtube-telegram-bot.git
cd youtube-telegram-bot
```
---

### 2️⃣ Configure config.yml
Create a `config.yml` file and fill in the details:
```yml
telegram_token: "YOUR_TELEGRAM_BOT_TOKEN"
chat_id: YOUR_TELEGRAM_CHAT_ID
youtube_api_key: "YOUR_YOUTUBE_API_KEY"
channels:
  - "UCe5Dq2HfS7IbF67qPnAuA5w"
  - "UCrYZanw_GNn7Q1bzVQEetTw"
  - "NEW_CHANNEL_ID"
```
- telegram_token – Bot token from BotFather
- chat_id – ID of the channel or group where the bot should send notifications
- youtube_api_key – YouTube API key
- channels – List of YouTube channels to monitor

🔹 How to find chat_id? – Use the bot @get_id_bot.

---

### 🚀 Running the Bot
🏗️ Locally (Go)

Install dependencies:

    go mod tidy

Run the bot:

    go run main.go

---

### 🐳 Docker

If you want to run the bot in a Docker container:

Build and start the container

    docker-compose up --build -d

Check logs

    docker logs -f youtube-telegram-bot

Stop the bot

    docker-compose down
---

### 🔄 Updating YouTube Channels

To add new YouTube channels for monitoring:

Edit `config.yml`
Save the file – the bot will automatically reload the new configuration!

---

### 🔧 Troubleshooting

If the bot isn’t working properly:

Check the logs:

    docker logs -f youtube-telegram-bot

Ensure `config.yml` is properly formatted
Make sure the bot has access to the Telegram channel (it must be an admin)
Verify that your YouTube API key is active

---

### 📜 License

This project is licensed under the MIT License. Feel free to modify and share it.

📬 If you have any questions, open an Issue or message me on Telegram! 🚀