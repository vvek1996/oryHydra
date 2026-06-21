

function LogoutCallback() {
  const login = () => {
    const authUrl = 'http://localhost:8080/.ory/hydra/oauth2/auth?client_id=36d0db37-f52e-46b6-bf1d-3923fc9cf46d&response_type=code&scope=openid&redirect_uri=http://localhost:8080/callback&state=securestate123';
    window.location.href = authUrl;
  };

  return (
    <div style={{
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      height: "100vh",
      fontFamily: "system-ui, -apple-system, sans-serif",
      backgroundColor: "#f5f5f5"
    }}>
      <div style={{
        backgroundColor: "#fff",
        padding: "40px",
        borderRadius: "8px",
        boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
        textAlign: "center",
        maxWidth: "400px"
      }}>
        <h2 style={{ color: "#333", marginBottom: "16px" }}>Logged Out</h2>
        <p style={{ color: "#666", marginBottom: "24px" }}>
          You have successfully logged out of your session.
        </p>
        <button 
          onClick={login}
          style={{
            backgroundColor: "#007bff",
            color: "#fff",
            border: "none",
            padding: "10px 20px",
            fontSize: "16px",
            borderRadius: "4px",
            cursor: "pointer",
            transition: "background-color 0.2s"
          }}
          onMouseOver={(e) => (e.currentTarget.style.backgroundColor = "#0056b3")}
          onMouseOut={(e) => (e.currentTarget.style.backgroundColor = "#007bff")}
        >
          Log In Again
        </button>
      </div>
    </div>
  );
}

export default LogoutCallback;
