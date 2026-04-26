-- Tự động tạo database etherpad khi MariaDB khởi động lần đầu
CREATE DATABASE IF NOT EXISTS etherpad CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
