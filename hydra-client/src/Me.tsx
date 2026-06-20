import { useEffect, useState } from "react";

function Me() {
  const [email, setEmail] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    // Retrieve the access token from localStorage. 
    // Make sure your callback route saves the token here using:
    // localStorage.setItem("access_token", data.access_token);
    const accessToken = localStorage.getItem("access_token");

    if (!accessToken) {
      setError("No access token found. Please log in first.");
      setLoading(false);
      return;
    }

    // Call your Express backend's protected /me endpoint
    fetch("http://localhost:4000/me", {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${accessToken}`,
        "Content-Type": "application/json",
      },
    })
      .then((res) => {
        if (!res.ok) {
          throw new Error("Token might be expired or invalid.");
        }
        return res.json();
      })
      .then((data) => {
        setEmail(data.email);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        setError("Failed to fetch user details. Please try logging in again.");
        setLoading(false);
      });
  }, []);

  const handleLogout = () => {
    const idToken = localStorage.getItem("id_token") || "";
    localStorage.removeItem("access_token");
    localStorage.removeItem("id_token");

    // Redirect to Hydra logout endpoint
    const hydraLogoutUrl = `http://localhost:4444/oauth2/sessions/logout?id_token_hint=${encodeURIComponent(
      idToken
    )}&post_logout_redirect_uri=${encodeURIComponent(
      "http://localhost:3000/logout-callback"
    )}`;
    window.location.href = hydraLogoutUrl;
  };

  if (loading) return <h2>Loading profile...</h2>;
  if (error) return <div style={{ color: "red", padding: "20px" }}>{error}</div>;

  return (
    <div style={{ padding: "20px", fontFamily: "system-ui, -apple-system, sans-serif" }}>
      <h2>My Profile</h2>
      <p><strong>Email:</strong> {email}</p>
      <button 
        onClick={handleLogout}
        style={{
          backgroundColor: "#dc3545",
          color: "#fff",
          border: "none",
          padding: "8px 16px",
          fontSize: "14px",
          borderRadius: "4px",
          cursor: "pointer",
          marginTop: "16px",
          transition: "background-color 0.2s"
        }}
        onMouseOver={(e) => (e.currentTarget.style.backgroundColor = "#c82333")}
        onMouseOut={(e) => (e.currentTarget.style.backgroundColor = "#dc3545")}
      >
        Logout
      </button>
    </div>
  );
}

export default Me;