CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    real_name VARCHAR(50) NOT NULL,
    role ENUM('employee', 'technician', 'admin') NOT NULL DEFAULT 'employee',
    status ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS devices (
    id INT AUTO_INCREMENT PRIMARY KEY,
    device_code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    model VARCHAR(100),
    location VARCHAR(200),
    purchase_date DATE,
    warranty_expire_date DATE,
    status ENUM('active', 'maintenance', 'scrapped') NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_devices_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS work_orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    device_id INT NOT NULL,
    employee_id INT NOT NULL,
    technician_id INT,
    fault_type ENUM('hardware', 'software', 'network', 'other') NOT NULL,
    fault_description TEXT NOT NULL,
    urgency ENUM('low', 'medium', 'high', 'urgent') NOT NULL DEFAULT 'low',
    status ENUM('pending_assign', 'assigned', 'processing', 'pending_confirm', 'closed') NOT NULL DEFAULT 'pending_assign',
    repair_measures TEXT,
    replaced_parts TEXT,
    repair_duration INT COMMENT '维修耗时，单位：分钟',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_work_orders_deleted_at (deleted_at),
    FOREIGN KEY (device_id) REFERENCES devices(id),
    FOREIGN KEY (employee_id) REFERENCES users(id),
    FOREIGN KEY (technician_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS images (
    id INT AUTO_INCREMENT PRIMARY KEY,
    work_order_id INT DEFAULT NULL,
    image_type ENUM('before', 'after') NOT NULL DEFAULT 'before',
    file_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_images_deleted_at (deleted_at),
    FOREIGN KEY (work_order_id) REFERENCES work_orders(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS operation_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    work_order_id INT NOT NULL,
    user_id INT NOT NULL,
    operation ENUM('create', 'assign', 'accept', 'process', 'submit', 'confirm', 'reject') NOT NULL,
    old_status ENUM('pending_assign', 'assigned', 'processing', 'pending_confirm', 'closed'),
    new_status ENUM('pending_assign', 'assigned', 'processing', 'pending_confirm', 'closed'),
    remark TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_operation_logs_deleted_at (deleted_at),
    FOREIGN KEY (work_order_id) REFERENCES work_orders(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- password: 123456
INSERT INTO users (username, password, real_name, role, status, phone) VALUES 
('admin', '$2a$10$dGcsllk.MEE8Np4llafjturWicVDnZ/7X81.kXXPa877dTQgYZ.Ky', '系统管理员', 'admin', 'active', '13800138000'),
('employee1', '$2a$10$dGcsllk.MEE8Np4llafjturWicVDnZ/7X81.kXXPa877dTQgYZ.Ky', '张三', 'employee', 'active', '13800138001'),
('tech1', '$2a$10$dGcsllk.MEE8Np4llafjturWicVDnZ/7X81.kXXPa877dTQgYZ.Ky', '李四', 'technician', 'active', '13800138002');

INSERT INTO devices (device_code, name, model, location, purchase_date, warranty_expire_date) VALUES 
('DEV001', '台式电脑', 'Dell OptiPlex 7010', '办公区A-101', '2022-01-15', '2025-01-15'),
('DEV002', '笔记本电脑', 'Lenovo ThinkPad X1 Carbon', '办公区B-201', '2023-03-20', '2026-03-20'),
('DEV003', '打印机', 'HP LaserJet Pro MFP M428fdw', '办公区A-102', '2021-06-10', '2024-06-10'),
('DEV004', '网络交换机', 'Cisco Catalyst 2960', '机房-1', '2020-11-05', '2023-11-05');
