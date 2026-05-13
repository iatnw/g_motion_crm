import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  allowedDevOrigins: ["gmotioncrm.com"],
  devIndicators: false,
};

export default nextConfig;
