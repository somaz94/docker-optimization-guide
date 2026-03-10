package com.example.bookapi;

import java.time.LocalDateTime;

public record Book(
    Long id,
    String title,
    String author,
    String isbn,
    int year,
    LocalDateTime createdAt
) {}
