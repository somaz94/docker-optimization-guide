// index.js
const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const { v4: uuidv4 } = require("uuid");

const app = express();

// Middleware
app.use(helmet());
app.use(cors());
app.use(morgan("combined"));
app.use(express.json());

// In-memory store
const tasks = [
  { id: uuidv4(), title: "Learn Docker", status: "done", priority: "high", createdAt: new Date().toISOString() },
  { id: uuidv4(), title: "Optimize images", status: "in-progress", priority: "high", createdAt: new Date().toISOString() },
];

// Routes
app.get("/health", (req, res) => {
  res.json({
    status: "healthy",
    uptime: process.uptime(),
    timestamp: new Date().toISOString(),
    memoryUsage: process.memoryUsage(),
  });
});

app.get("/tasks", (req, res) => {
  const { status, priority } = req.query;
  let filtered = tasks;
  if (status) filtered = filtered.filter((t) => t.status === status);
  if (priority) filtered = filtered.filter((t) => t.priority === priority);
  res.json({ status: "success", count: filtered.length, data: filtered });
});

app.get("/tasks/:id", (req, res) => {
  const task = tasks.find((t) => t.id === req.params.id);
  if (!task) return res.status(404).json({ status: "error", message: "Task not found" });
  res.json({ status: "success", data: task });
});

app.post("/tasks", (req, res) => {
  const { title, priority = "medium" } = req.body;
  if (!title) return res.status(400).json({ status: "error", message: "Title is required" });

  const task = {
    id: uuidv4(),
    title,
    status: "todo",
    priority,
    createdAt: new Date().toISOString(),
  };
  tasks.push(task);
  res.status(201).json({ status: "success", data: task });
});

app.patch("/tasks/:id", (req, res) => {
  const task = tasks.find((t) => t.id === req.params.id);
  if (!task) return res.status(404).json({ status: "error", message: "Task not found" });

  const { title, status, priority } = req.body;
  if (title) task.title = title;
  if (status) task.status = status;
  if (priority) task.priority = priority;
  res.json({ status: "success", data: task });
});

app.delete("/tasks/:id", (req, res) => {
  const index = tasks.findIndex((t) => t.id === req.params.id);
  if (index === -1) return res.status(404).json({ status: "error", message: "Task not found" });
  tasks.splice(index, 1);
  res.json({ status: "success", message: "Task deleted" });
});

const port = process.env.PORT || 3000;
app.listen(port, () => {
  console.log(`Task API server running on port ${port}`);
});
