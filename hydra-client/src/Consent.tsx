import { useEffect, useRef } from "react";

function Consent() {
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const params = new URLSearchParams(window.location.search);
    const challenge = params.get("consent_challenge");

    if (!challenge) return;

     window.location.href = `http://localhost:4000/consent?consent_challenge=${challenge}`;
     
    // fetch(`http://localhost:4000/consent?consent_challenge=${challenge}`)
    //   .then(res => res.json())
    //   .then(data => {
    //     window.location.href = data.redirect_to;
    //   });
  }, []);

  return <h2>Giving consent...</h2>;
}

export default Consent;