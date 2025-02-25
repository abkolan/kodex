CREATE TABLE
    auction (
        id INT PRIMARY KEY AUTO_INCREMENT,
        listing_id INT NOT NULL,
        max_bid_id INT,
        max_bid_amount INT DEFAULT 0 NOT NULL
    );

CREATE TABLE
    bids (
        id INT PRIMARY KEY AUTO_INCREMENT,
        auction_id INT NOT NULL,
        amount DECIMAL(10, 2) NOT NULL,
        user_id INT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        status ENUM ('accepted', 'rejected') NOT NULL
    );