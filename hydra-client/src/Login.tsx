import { useEffect, useRef } from "react";

function Login() {
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const params = new URLSearchParams(window.location.search);
    const challenge = params.get("login_challenge");

    if (!challenge) return;

    fetch(`http://localhost:4000/login?login_challenge=${challenge}`)
      .then(res => res.json())
      .then(data => {
        window.location.href = data.redirect_to;
      });
  }, []);

  return <h2>Logging in...</h2>;
}

export default Login;