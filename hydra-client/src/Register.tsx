import { useEffect, useState } from "react";
import axios from "axios";

const getCsrfToken = (flow: any) => {
  const node = flow.ui.nodes.find(
    (n: any) => n.attributes.name === "csrf_token"
  );
  
  return node?.attributes.value;
};

function Register() {
  const [flow, setFlow] = useState<any>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const flowId = params.get("flow");

    if (!flowId) return;

    // Fetch registration flow
    axios
      .get(
        `http://localhost:8080/.ory/kratos/self-service/registration/flows?id=${flowId}`,
        {
          withCredentials: true // VERY IMPORTANT
        }
      )
      .then((res) => {
        setFlow(res.data);
      })
      .catch((err) => {
        console.error("Flow error:", err);
      });
  }, []);

  const submit = async () => {
    try {
      const csrfToken = getCsrfToken(flow);
      await axios.post(
        flow.ui.action,
        {
          method: "password",
          password: password,
          csrf_token: csrfToken,
          traits: {
            email: email
          }
        },
        {
          headers: {
            "Content-Type": "application/json"
          },
          withCredentials: true // ADD THIS HERE
        }
      );

      alert("🎉 Registration successful");

      // redirect to login
      window.location.href = "/";
    } catch (err: any) {
      console.error("Register error:", err.response?.data);
      alert("Registration failed ❌");
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>Register</h2>
      
      <input
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <br /><br />
      
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <br /><br />
      
      <button onClick={submit}>Register</button>
    </div>
  );
}

export default Register;