import { useState } from "react";
import styles from "./Register.module.css";
import { registerUser } from "../customHooks/api";
import { useNavigate } from "react-router-dom";

const Register = () => {
  const [email, setEmail] = useState("");
  const [nama, setNama] = useState("");
  const [password, setPassword] = useState("");
  const [isError, setIsError] = useState(false);
  const [infoError, setInfoError] = useState("");
  const navigate = useNavigate();
  const handleRegister = async () => {
    try {
      await registerUser({ email, nama, password });
      alert("Registrasi berhasil!");
      setIsError(false);
      navigate("/login");
    } catch (error) {
      setIsError(true);
      setInfoError(`${error}`);
    }
  };

  return (
    <div className="container">
      <h2 className={styles.titleRegister}>Register</h2>
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        className={styles.inputReg}
      />
      <input
        type="text"
        placeholder="Nama"
        value={nama}
        onChange={(e) => setNama(e.target.value)}
        className={styles.inputReg}
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        className={styles.inputReg}
      />
      <button onClick={handleRegister} className={styles.btnReg}>
        Register
      </button>
      {isError && <p className={styles.errorText}>{infoError}</p>}
    </div>
  );
};

export default Register;
