const express = require('express');
const app = express();

app.use(express.json());

// --- [🛡️ หน้าแรก: ป้องกัน 404 และเช็คสถานะระบบ] ---
app.get('/', (req, res) => {
  res.status(200).send(`
    <html>
      <body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
        <h1 style="color: #0F9D58;">🐅 ThitNueaHub V7 Ignite Online!</h1>
        <p>ระบบทำงานปกติในบ้านหลังใหม่ (Mobile AI Lab) เรียบร้อยแล้วค่ะบอส</p>
        <div style="margin-top: 20px; color: #666;">
          <small>Project ID: thitnueahub-mobile-ai-lab</small>
        </div>
      </body>
    </html>
  `);
});

// --- [🤖 Webhook สำหรับรับค่าจาก LINE หรือ API อื่นๆ] ---
app.post('/webhook', (req, res) => {
  console.log('--- Received Webhook Data ---');
  console.log(JSON.stringify(req.body, null, 2));
  res.status(200).send('OK');
});

// --- [🔌 Port Setup สำหรับ Google Cloud Run] ---
const port = process.env.PORT || 8080;
app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
