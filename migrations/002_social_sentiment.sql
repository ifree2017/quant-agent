-- Social Sentiment Tables
-- 社交情绪数据表

CREATE TABLE IF NOT EXISTS social_sentiment (
    id SERIAL PRIMARY KEY,
    stock_code VARCHAR(20) NOT NULL,
    platform VARCHAR(20) NOT NULL,  -- 'eastmoney' | 'xueqiu'
    sentiment_score DECIMAL(5,2),  -- -100 ~ +100
    post_count INT,
    comment_count INT,
    heat_score DECIMAL(10,2),
    keywords TEXT[],
    fetched_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS social_posts (
    id SERIAL PRIMARY KEY,
    stock_code VARCHAR(20) NOT NULL,
    platform VARCHAR(20) NOT NULL,  -- 'eastmoney' | 'xueqiu'
    title VARCHAR(500),
    content TEXT,
    author VARCHAR(100),
    likes INT DEFAULT 0,
    comments INT DEFAULT 0,
    sentiment DECIMAL(5,2),
    publish_time TIMESTAMP,
    fetched_at TIMESTAMP DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_social_sentiment_stock ON social_sentiment(stock_code);
CREATE INDEX IF NOT EXISTS idx_social_sentiment_platform ON social_sentiment(platform);
CREATE INDEX IF NOT EXISTS idx_social_sentiment_fetched ON social_sentiment(fetched_at);
CREATE INDEX IF NOT EXISTS idx_social_posts_stock ON social_posts(stock_code);
CREATE INDEX IF NOT EXISTS idx_social_posts_platform ON social_posts(platform);
CREATE INDEX IF NOT EXISTS idx_social_posts_publish ON social_posts(publish_time);
