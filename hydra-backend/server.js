const express = require("express");
const axios = require("axios");
const cors = require("cors");

const app = express();
app.use(express.json());
// Use specific CORS configuration for better security, allowing credentials (cookies)
// to be sent from your frontend origin.
app.use(cors({ origin: 'http://localhost:3000', credentials: true }));

/**
 * LOGIN ACCEPT
 */
app.get("/login", async (req, res) => {
  const challenge = req.query.login_challenge;

  try {
    // Check for an active session with Ory Kratos using the incoming cookies
    const sessionResponse = await axios.get("http://localhost:4433/sessions/whoami", {
      headers: { Cookie: req.headers.cookie || "" }
    });
    const subjectId = sessionResponse.data.identity.id;

    const response = await axios.put(
      `http://localhost:4445/oauth2/auth/requests/login/accept?login_challenge=${challenge}`,
      {
        subject: subjectId,
        remember: true,
        remember_for: 3600,
      }
    );

    // Redirect the browser back to Hydra to continue the OAuth2 flow
    res.redirect(response.data.redirect_to);
  } catch (err) {
    if (err.response && err.response.status === 401) {
      // Real-world scenario: User has no session. Redirect to the frontend login page.
      // Pass the login_challenge so the frontend knows where to return after login.
      return res.redirect(`http://localhost:3000/login?login_challenge=${challenge}`);
    }

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
    // It's crucial to re-validate the session here. The user might have logged out
    // in another tab or their session might have expired.
    if (!req.headers.cookie) {
      throw { response: { status: 401 } };
    }

    // Fetch the user's active Kratos session using their cookies to get the email
    const sessionResponse = await axios.get("http://localhost:4433/sessions/whoami", {
      headers: { Cookie: req.headers.cookie || "" }
    });
    const email = sessionResponse.data.identity.traits.email;

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
        session: {
          id_token: {
            email: email
          },
          access_token: {
            email: email
          }
        }
      }
    );

    // Redirect the browser back to Hydra to complete the consent flow
    res.redirect(response.data.redirect_to);
  } catch (err) {
    // This block handles cases where the Kratos session is invalid or expired.
    if (err.response && err.response.status === 401) {
      // The user is not logged in, so we redirect them to the login page,
      // passing the consent_challenge along so we can resume the flow after login.
      return res.redirect(`http://localhost:3000/login?consent_challenge=${challenge}`);
    }
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

/**
 * PROTECTED API ENDPOINT (/me)
 * Validates the Access Token and returns the user's email
 */
app.get("/me", async (req, res) => {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    return res.status(401).json({ error: "Missing Bearer token" });
  }

  const token = authHeader.split(" ")[1];

  try {
    // Introspect the token with Hydra Admin API to check if it is active
    const introspectResponse = await axios.post(
      "http://localhost:4445/oauth2/introspect",
      new URLSearchParams({ token: token }),
      { headers: { "Content-Type": "application/x-www-form-urlencoded" } }
    );

    const data = introspectResponse.data;

    if (!data.active) {
      return res.status(401).json({ error: "Token is inactive or expired" });
    }

    res.json({
      // Hydra exposes session.access_token payload inside the 'ext' object
      email: data.ext?.email || data.email,
      subject: data.sub
    });
  } catch (err) {
    console.error(err.response?.data || err);
    res.status(500).json({ error: "Failed to validate token" });
  }
});

/**
 * LOGOUT ENDPOINT
 * Handles the Hydra logout challenge, revokes Kratos session, and completes the flow
 */
app.get("/logout", async (req, res) => {
  const challenge = req.query.logout_challenge;

  // If there's no challenge, redirect to frontend home
  if (!challenge) {
    return res.redirect("http://localhost:3000/");
  }

  try {
    // 1. Accept the logout challenge in Ory Hydra
    const response = await axios.put(
      `http://localhost:4445/oauth2/auth/requests/logout/accept?logout_challenge=${challenge}`
    );
    const redirectUrl = response.data.redirect_to;

    // 2. Fetch browser logout URL from Ory Kratos
    try {
      const kratosResponse = await axios.get(
        "http://localhost:4433/self-service/logout/browser",
        {
          headers: { Cookie: req.headers.cookie || "" }
        }
      );
      const kratosLogoutUrl = kratosResponse.data.logout_url;
      // Redirect to Kratos logout URL, passing the Hydra redirect_to as return_to
      return res.redirect(
        `${kratosLogoutUrl}&return_to=${encodeURIComponent(redirectUrl)}`
      );
    } catch (kratosErr) {
      // If the Kratos session is invalid/expired (e.g. 401), we just complete Hydra logout directly
      console.warn("Kratos logout failed or no active session:", kratosErr.response?.data || kratosErr.message);
      return res.redirect(redirectUrl);
    }
  } catch (err) {
    console.error(err.response?.data || err);
    res.status(500).send("Logout failed");
  }
});

app.listen(4000, () => {
  console.log("Backend running on http://localhost:4000");
});