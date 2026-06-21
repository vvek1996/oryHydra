import { useEffect, useState, useRef } from "react";

function Callback() {
  const [message, setMessage] = useState("Processing login...");
  const [tokens, setTokens] = useState<any>(null);
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const params = new URLSearchParams(window.location.search);
    const code = params.get("code");
    const error = params.get("error");

    if (error) {
      setMessage(`OAuth error: ${error}`);
      return;
    }

    if (!code) {
      setMessage("No authorization code found ❌");
      return;
    }

    console.log("📝 Code received:", code);

    // 📝 Clean URL to avoid reuse
    window.history.replaceState({}, document.title, "/callback");

    fetch("http://localhost:8080/api/token", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ code })
    })
      .then(res => res.json())
      .then(data => {
        console.log("📝 Tokens:", data);

        if (data.error) {
          setMessage("Token exchange failed ❌");
          return;
        }

        // ✅ SAVE THE TOKEN HERE
        if (data.access_token) {
          localStorage.setItem("access_token", data.access_token);
          if (data.id_token) {
            localStorage.setItem("id_token", data.id_token);
          }
          
          // Automatically redirect to the /me page
          window.location.href = "/me";
        }

        setTokens(data);
        setMessage("Login successful 🎉");
      })
      .catch(err => {
        console.error("❌ Fetch error:", err);
        setMessage("Token exchange failed ❌");
      });
  }, []);

  return (
    <div style={{ padding: 20 }}>
      <h2>Callback Page</h2>
      
      <p>{message}</p>

      {tokens && (
        <pre style={{
          background: "#F5F5F5",
          padding: 10,
          borderRadius: 5
        }}>
          {JSON.stringify(tokens, null, 2)}
        </pre>
      )}
    </div>
  );
}

export default Callback;