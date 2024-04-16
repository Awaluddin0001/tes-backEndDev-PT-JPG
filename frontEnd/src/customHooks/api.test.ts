import axios from "axios";
import { registerUser } from "./api";
import { describe, it, expect, beforeEach, vi, Mock } from "vitest";
vi.mock("axios");

describe("API Tests", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should register a user successfully", async () => {
    // Arrange
    const userData = {
      email: "test@example.com",
      nama: "Test User",
      password: "password123",
    };
    const responseData = { id: 1, ...userData };
    (axios.post as Mock).mockResolvedValue({ data: responseData });

    // Act
    const result = await registerUser(userData);

    // Assert
    expect(result).toEqual(responseData);
    expect(axios.post).toHaveBeenCalledWith(
      "http://localhost:3000/user",
      userData
    );
  });

  it("should throw an error when registration fails", async () => {
    // Arrange
    const userData = {
      email: "test@example.com",
      nama: "Test User",
      password: "password123",
    };
    const errorMessage = "Registration failed";
    (axios.post as Mock).mockRejectedValue(new Error(errorMessage));

    // Act & Assert
    await expect(registerUser(userData)).rejects.toThrow(errorMessage);
    expect(axios.post).toHaveBeenCalledWith(
      "http://localhost:3000/user",
      userData
    );
  });

  it("should throw an error if email already exists", async () => {
    // Arrange
    const userData = {
      email: "test@example.com",
      nama: "Test User",
      password: "password123",
    };
    const errorMessage = "Email already exists";
    (axios.post as Mock).mockRejectedValue(new Error(errorMessage));

    // Act & Assert
    await expect(registerUser(userData)).rejects.toThrow(errorMessage);
    expect(axios.post).toHaveBeenCalledWith(
      "http://localhost:3000/user",
      userData
    );
  });

  it("should throw an error if email is not unique", async () => {
    // Arrange
    const userData = {
      email: "test@example.com",
      nama: "Test User",
      password: "password123",
    };
    const errorMessage = "Email already exists";
    const spy = vi
      .spyOn(axios, "post")
      .mockRejectedValueOnce(new Error(errorMessage)) // Panggilan pertama gagal
      .mockResolvedValueOnce({ data: { id: 2, ...userData } }); // Panggilan kedua berhasil

    // Act & Assert
    await expect(registerUser(userData)).rejects.toThrow(errorMessage);
    await expect(registerUser(userData)).resolves.toEqual({
      id: 2,
      ...userData,
    }); // Harus berhasil
    expect(spy).toHaveBeenCalledTimes(2);
  });
});
