const express = require("express");
const axios = require("axios");
const cors = require("cors");

const app = express();
app.use(express.json());
app.use(cors());

/**
 * LOGIN ACCEPT
 */
app.get("/login", async (req, res) => {
  const challenge = req.query.login_challenge;

  try {
    const response = await axios.put(
      `http://localhost:4445/oauth2/auth/requests/login/accept?login_challenge=${challenge}`,
      {
        subject: "demo-user",
        remember: true,
        remember_for: 3600,
      }
    );

    res.json(response.data);
  } catch (err) {
    console.error(err.response?.data || err);
    res.status(500).send("Login failed");
  }
});

/**
 * CONSENT ACCEPT
 */
app.get("/consent", async (req, res) => {
  const challenge = req.query.consent_challenge;

  try {
    // Step 1: GET consent request details
    const consentReq = await axios.get(
      `http://localhost:4445/oauth2/auth/requests/consent?consent_challenge=${challenge}`
    );

    const requestedScopes = consentReq.data.requested_scope;

    // Step 2: ACCEPT consent properly
    const response = await axios.put(
      `http://localhost:4445/oauth2/auth/requests/consent/accept?consent_challenge=${challenge}`,
      {
        grant_scope: requestedScopes,
        remember: true,
        remember_for: 3600,
      }
    );

    res.json(response.data);
  } catch (err) {
    console.error(err.response?.data || err);
    res.status(500).send("Consent failed");
  }
});

app.post("/token", async (req, res) => {
  const { code } = req.body;

  try {
    const response = await axios.post(
      "http://localhost:4444/oauth2/token",
      new URLSearchParams({
        grant_type: "authorization_code",
        code: code,
        redirect_uri: "http://localhost:3000/callback",
      }),
      {
        auth: {
          username: "36d0db37-f52e-46b6-bf1d-3923fc9cf46d",
          password: "secret",
        },
      }
    );

    res.json(response.data);
  } catch (err) {
    console.error(err.response?.data || err);
    res.status(500).send("Token exchange failed");
  }
});

app.listen(4000, () => {
  console.log("Backend running on http://localhost:4000");
});