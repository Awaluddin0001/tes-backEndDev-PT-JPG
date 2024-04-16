import { useState, useEffect } from "react";
import styles from "./Dashboard.module.css";
import { addSales } from "../customHooks/api";
import { useAuth } from "../customHooks/AuthContext";
import { useNavigate } from "react-router-dom";
import LogOut from "../components/LogOut";
import { formatDateToISO } from "../customHooks/TimeConvert";
const Dashboard = () => {
  const [tanggal, setTanggal] = useState("");
  const [jenis, setJenis] = useState("Barang");
  const [nominal, setNominal] = useState(0);
  const [nominalShow, setNominalShow] = useState("");
  const { accessToken } = useAuth();
  const navigate = useNavigate();
  useEffect(() => {
    if (!accessToken) {
      navigate("/login");
    }
  }, [accessToken]);

  const handleInputSales = async () => {
    try {
      const tanggalGolangFormat = formatDateToISO(tanggal);
      await addSales({
        tanggal: tanggalGolangFormat,
        jenis,
        nominal,
        accessToken,
      });
      alert("Input penjualan berhasil!");
      setTanggal("");
      setJenis("Barang");
      setNominal(0);
      setNominalShow("");
    } catch (error) {
      alert("Gagal memasukkan penjualan. Mohon coba lagi.");
    }
  };

  return (
    <div className="container">
      <LogOut />
      <h2 className={styles.titleDash}>Input Penjualan</h2>
      <input
        type="date"
        placeholder="Tanggal"
        value={tanggal}
        onChange={(e) => setTanggal(e.target.value)}
        className={styles.inputDash}
      />
      <select
        value={jenis}
        onChange={(e) => setJenis(e.target.value)}
        className={styles.inputDash}
      >
        <option value="Barang" className={styles.inputDash}>
          Barang
        </option>
        <option value="Jasa" className={styles.inputDash}>
          Jasa
        </option>
      </select>
      <input
        type="number"
        placeholder="Nominal"
        value={nominalShow}
        onChange={(e) => {
          setNominal(Number(e.target.value));
          setNominalShow(e.target.value);
        }}
        className={styles.inputDash}
      />
      <button onClick={handleInputSales} className={styles.btnDash}>
        Submit
      </button>
    </div>
  );
};

export default Dashboard;
