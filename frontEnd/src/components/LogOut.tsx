import { useAuth } from "../customHooks/AuthContext";
import { useNavigate } from "react-router-dom";
export default function LogOut() {
  const { setAccessToken } = useAuth();
  const navigate = useNavigate();
  const handleInputSales = () => {
    navigate("/dashboard");
  };
  const handleReportSales = () => {
    navigate("/reportsales");
  };
  const handleLogOut = () => {
    setAccessToken(null);
  };

  return (
    <div
      style={{
        position: "fixed",
        top: "20px",
        right: "160px",
      }}
    >
      <div style={{ display: "flex", flexDirection: "row", gap: "14px" }}>
        <button onClick={handleInputSales} style={{ fontSize: "14rem" }}>
          Input Sales
        </button>
        <button onClick={handleReportSales} style={{ fontSize: "14rem" }}>
          Report Sales
        </button>
        <button onClick={handleLogOut} style={{ fontSize: "14rem" }}>
          Logout
        </button>
      </div>
    </div>
  );
}
