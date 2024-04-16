import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import styles from "./Login.module.css";
import { loginUser } from "../customHooks/api";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../customHooks/AuthContext";

const Login = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isError, setIsError] = useState(false);
  const [infoError, setInfoError] = useState("");
  const navigate = useNavigate();
  const { accessToken, setAccessToken } = useAuth();

  useEffect(() => {
    if (accessToken) {
      navigate("/dashboard");
    }
  }, [accessToken]);

  const handleLogin = async () => {
    try {
      const user = await loginUser({ email, password });
      setAccessToken(user.token);
      alert("Login berhasil!");
      setIsError(false);
      navigate("/dashboard");
      // Redirect ke halaman lain atau lakukan sesuatu setelah login berhasil
    } catch (error) {
      setIsError(true);
      setInfoError(`${error}`);
    }
  };

  return (
    <div className="container">
      <h1 className={styles.titleLogin}>
        Back Office Sale PT Jaya Pratama Groserindo
      </h1>
      <h2 className={styles.subtitleLogin}>Login</h2>
      {isError && <p className={styles.errorText}>{infoError}</p>}
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        className={styles.inputLogin}
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        className={styles.inputLogin}
      />
      <button className={styles.btnLogin} onClick={handleLogin}>
        Login
      </button>
      <Link to={"/register"} className={styles.regButton}>
        Belum punya Akun ? klik disini
      </Link>
    </div>
  );
};

export default Login;
