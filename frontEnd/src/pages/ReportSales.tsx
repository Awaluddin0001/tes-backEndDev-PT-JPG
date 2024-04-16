import { useState, useEffect } from "react";
import LogOut from "../components/LogOut";
import styles from "./ReportSales.module.css";
import { generateReport } from "../customHooks/api";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../customHooks/AuthContext";
const ReportSales = () => {
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const navigate = useNavigate();
  const { accessToken } = useAuth();
  useEffect(() => {
    if (!accessToken) {
      navigate("/login");
    }
  }, [accessToken]);

  const handleGenerateReport = async () => {
    try {
      await generateReport(startDate, endDate, accessToken);
      alert("Laporan berhasil diunduh.");
    } catch (error) {
      alert("Gagal mengunduh laporan. Mohon coba lagi.");
    }
  };

  return (
    <div className="container">
      <LogOut />
      <h2 className={styles.titleReport}>Report Penjualan</h2>
      <h3 className={styles.subtitleReport}>Tanggal Awal</h3>
      <input
        type="date"
        placeholder="Start Date"
        value={startDate}
        onChange={(e) => setStartDate(e.target.value)}
        className={styles.inputRep}
      />
      <h3 className={styles.subtitleReport}>Tanggal Akhir</h3>
      <input
        type="date"
        placeholder="End Date"
        value={endDate}
        onChange={(e) => setEndDate(e.target.value)}
        className={styles.inputRep}
      />
      <button onClick={handleGenerateReport} className={styles.btnRep}>
        Download Report
      </button>
    </div>
  );
};

export default ReportSales;
