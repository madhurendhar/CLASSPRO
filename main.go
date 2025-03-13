import os
import time
import requests
from datetime import datetime, timedelta
from flask import Flask, jsonify
from flask_mail import Mail, Message
from apscheduler.schedulers.background import BackgroundScheduler

app = Flask(__name__)

# Email Configurations
app.config['MAIL_SERVER'] = 'smtp.gmail.com'
app.config['MAIL_PORT'] = 587
app.config['MAIL_USE_TLS'] = True
app.config['MAIL_USERNAME'] = os.getenv('EMAIL_USER')
app.config['MAIL_PASSWORD'] = os.getenv('EMAIL_PASS')
mail = Mail(app)

# Data storage (Mock Database)
exam_assignments = []

# Function to scrape timetable for exams/assignments
def fetch_exam_assignments():
    url = "https://srmist.edu.in/timetable"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        exam_assignments.clear()
        for item in data.get('timetable', []):
            if 'exam' in item['event'].lower() or 'assignment' in item['event'].lower():
                exam_assignments.append({
                    'subject': item['subject'],
                    'event': item['event'],
                    'date': item['date'],
                    'time': item['time']
                })

# Function to send reminders
def send_reminders():
    now = datetime.now()
    for item in exam_assignments:
        event_date = datetime.strptime(item['date'], '%Y-%m-%d')
        if now + timedelta(days=1) >= event_date:  # Send reminder 1 day before
            msg = Message(
                subject=f"Reminder: {item['event']} for {item['subject']}",
                sender=app.config['MAIL_USERNAME'],
                recipients=['student@example.com'],  # Replace with actual student email
                body=f"Hey! Don't forget your {item['event']} for {item['subject']} on {item['date']} at {item['time']}."
            )
            mail.send(msg)

# Background Scheduler
scheduler = BackgroundScheduler()
scheduler.add_job(fetch_exam_assignments, 'interval', hours=12)
scheduler.add_job(send_reminders, 'interval', hours=24)
scheduler.start()

@app.route('/exam_assignments', methods=['GET'])
def get_exam_assignments():
    return jsonify(exam_assignments)

if __name__ == '__main__':
    fetch_exam_assignments()  # Initial Fetch
    app.run(debug=True)

