use std::error::Error;
use std::fmt;

#[derive(Debug)]
pub struct MyError {
    pub message: String,
}

impl fmt::Display for MyError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "错误: {}", self.message)
    }
}

impl Error for MyError {
    fn description(&self) -> &str {
        &self.message
    }
}
