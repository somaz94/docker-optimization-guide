package com.example.bookapi;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.atomic.AtomicLong;

@RestController
@RequestMapping("/api")
public class BookController {

    private final List<Book> books = new CopyOnWriteArrayList<>();
    private final AtomicLong idCounter = new AtomicLong(0);

    public BookController() {
        books.add(new Book(idCounter.incrementAndGet(), "Clean Code", "Robert C. Martin", "978-0132350884", 2008, LocalDateTime.now()));
        books.add(new Book(idCounter.incrementAndGet(), "The Pragmatic Programmer", "David Thomas", "978-0135957059", 2019, LocalDateTime.now()));
    }

    @GetMapping("/health")
    public Map<String, Object> health() {
        return Map.of(
            "status", "healthy",
            "timestamp", LocalDateTime.now().toString(),
            "service", "Book API",
            "javaVersion", System.getProperty("java.version")
        );
    }

    @GetMapping("/books")
    public Map<String, Object> getBooks(@RequestParam(required = false) String author) {
        List<Book> filtered = books;
        if (author != null) {
            filtered = books.stream()
                .filter(b -> b.author().toLowerCase().contains(author.toLowerCase()))
                .toList();
        }
        return Map.of("status", "success", "count", filtered.size(), "data", filtered);
    }

    @GetMapping("/books/{id}")
    public ResponseEntity<Map<String, Object>> getBook(@PathVariable Long id) {
        return books.stream()
            .filter(b -> b.id().equals(id))
            .findFirst()
            .map(book -> ResponseEntity.ok(Map.<String, Object>of("status", "success", "data", book)))
            .orElse(ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(Map.of("status", "error", "message", "Book not found")));
    }

    @PostMapping("/books")
    public ResponseEntity<Map<String, Object>> createBook(@RequestBody Map<String, Object> body) {
        String title = (String) body.get("title");
        String author = (String) body.get("author");
        String isbn = (String) body.getOrDefault("isbn", "");
        int year = (int) body.getOrDefault("year", 2024);

        if (title == null || author == null) {
            return ResponseEntity.badRequest()
                .body(Map.of("status", "error", "message", "Title and author are required"));
        }

        Book book = new Book(idCounter.incrementAndGet(), title, author, isbn, year, LocalDateTime.now());
        books.add(book);
        return ResponseEntity.status(HttpStatus.CREATED)
            .body(Map.of("status", "success", "data", book));
    }

    @DeleteMapping("/books/{id}")
    public ResponseEntity<Map<String, Object>> deleteBook(@PathVariable Long id) {
        boolean removed = books.removeIf(b -> b.id().equals(id));
        if (!removed) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(Map.of("status", "error", "message", "Book not found"));
        }
        return ResponseEntity.ok(Map.of("status", "success", "message", "Book deleted"));
    }
}
