import React from "react";
import { BrowserRouter, Routes, Route, useNavigate } from "react-router-dom";
import Callback from "./Callback";
import Login from "./Login";
import Register from "./Register";
import Consent from "./Consent";

function Home() {
  const login = () => {
    const authUrl = 'http://localhost:4444/oauth2/auth?client_id=36d0db37-f52e-46b6-bf1d-3923fc9cf46d&response_type=code&scope=openid&redirect_uri=http://localhost:3000/callback&state=securestate123';
    window.location.href = authUrl;
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>Hydra OAuth Client</h2>
      <button onClick={login}>Login with Hydra</button>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/callback" element={<Callback />} />

        {/* the below should be the authenticators like google ui, github ui */}
        <Route path="/login" element={<Login />} />
        <Route path="/consent" element={<Consent />} />
        <Route path="/register" element={<Register />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;