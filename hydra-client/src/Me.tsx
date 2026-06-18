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

  if (loading) return <h2>Loading profile...</h2>;
  if (error) return <div style={{ color: "red", padding: "20px" }}>{error}</div>;

  return (
    <div style={{ padding: "20px" }}>
      <h2>My Profile</h2>
      <p><strong>Email:</strong> {email}</p>
    </div>
  );
}

export default Me;