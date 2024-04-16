import axios from "axios";
const BASE_URL = "http://localhost:3000"; // Ganti dengan URL backend Anda

interface RegisterData {
  email: string;
  nama: string;
  password: string;
}

interface LoginData {
  email: string;
  password: string;
}

interface SalesData {
  tanggal: string;
  jenis: string;
  nominal: number;
  accessToken: string | null;
}

type ErrorWithMessage = {
  response: {
    data: string;
  };
};

function isErrorWithMessage(error: unknown): error is ErrorWithMessage {
  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof (error as ErrorWithMessage).response === "object" &&
    "data" in (error as ErrorWithMessage).response &&
    typeof (error as ErrorWithMessage).response.data === "string"
  ) {
    return true;
  }
  return false;
}

function toErrorWithMessage(maybeError: unknown): ErrorWithMessage {
  if (isErrorWithMessage(maybeError)) return maybeError;

  try {
    const { data } = (maybeError as ErrorWithMessage).response;
    if (typeof data === "string") {
      return { response: { data } };
    } else {
      throw new Error("Invalid data type");
    }
  } catch {
    return { response: { data: "" } };
  }
}

function getErrorMessage(error: unknown) {
  return toErrorWithMessage(error).response.data;
}

export const registerUser = async (userData: RegisterData) => {
  try {
    const response = await axios.post(`${BASE_URL}/user`, userData);
    return response.data;
  } catch (error) {
    throw getErrorMessage(error);
  }
};

export const loginUser = async (loginData: LoginData) => {
  try {
    const response = await axios.post(`${BASE_URL}/login`, loginData);
    return response.data;
  } catch (error) {
    throw getErrorMessage(error);
  }
};

export const addSales = async (salesData: SalesData) => {
  try {
    const response = await axios.post(`${BASE_URL}/sales`, salesData, {
      headers: {
        Authorization: `${salesData.accessToken}`,
      },
    });
    return response.data;
  } catch (error) {
    console.log(error);
    throw getErrorMessage(error);
  }
};

export const generateReport = async (
  startDate: string,
  endDate: string,
  accessToken: string | null
): Promise<void> => {
  try {
    if (!accessToken) {
      throw new Error("Token akses tidak tersedia");
    }

    const response = await axios.post(`${BASE_URL}/report`, null, {
      params: { start_date: startDate, end_date: endDate },
      responseType: "blob",
      headers: {
        Authorization: accessToken,
      },
    });

    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", "report.xlsx");
    document.body.appendChild(link);
    link.click();
  } catch (error) {
    console.log(error);
    console.error("Gagal mengambil laporan:", error);
    // Tambahkan logika penanganan kesalahan di sini
    // Misalnya, munculkan pesan kesalahan kepada pengguna
  }
};
