// main.rs
use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;
use std::sync::Mutex;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Note {
    id: Uuid,
    title: String,
    content: String,
    tags: Vec<String>,
    created_at: DateTime<Utc>,
}

#[derive(Debug, Deserialize)]
struct CreateNote {
    title: String,
    content: String,
    #[serde(default)]
    tags: Vec<String>,
}

#[derive(Serialize)]
struct ApiResponse<T: Serialize> {
    status: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<T>,
    #[serde(skip_serializing_if = "Option::is_none")]
    message: Option<String>,
}

struct AppState {
    notes: Mutex<Vec<Note>>,
}

async fn health() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "healthy",
        "timestamp": Utc::now().to_rfc3339(),
        "service": "Note API",
    }))
}

async fn get_notes(data: web::Data<AppState>, query: web::Query<std::collections::HashMap<String, String>>) -> HttpResponse {
    let notes = data.notes.lock().unwrap();
    let filtered: Vec<&Note> = if let Some(tag) = query.get("tag") {
        notes.iter().filter(|n| n.tags.contains(tag)).collect()
    } else {
        notes.iter().collect()
    };

    HttpResponse::Ok().json(ApiResponse {
        status: "success".to_string(),
        data: Some(&filtered),
        message: None,
    })
}

async fn get_note(data: web::Data<AppState>, path: web::Path<Uuid>) -> HttpResponse {
    let notes = data.notes.lock().unwrap();
    let id = path.into_inner();

    match notes.iter().find(|n| n.id == id) {
        Some(note) => HttpResponse::Ok().json(ApiResponse {
            status: "success".to_string(),
            data: Some(note.clone()),
            message: None,
        }),
        None => HttpResponse::NotFound().json(ApiResponse::<()> {
            status: "error".to_string(),
            data: None,
            message: Some("Note not found".to_string()),
        }),
    }
}

async fn create_note(data: web::Data<AppState>, body: web::Json<CreateNote>) -> HttpResponse {
    let note = Note {
        id: Uuid::new_v4(),
        title: body.title.clone(),
        content: body.content.clone(),
        tags: body.tags.clone(),
        created_at: Utc::now(),
    };

    data.notes.lock().unwrap().push(note.clone());

    HttpResponse::Created().json(ApiResponse {
        status: "success".to_string(),
        data: Some(note),
        message: None,
    })
}

async fn delete_note(data: web::Data<AppState>, path: web::Path<Uuid>) -> HttpResponse {
    let mut notes = data.notes.lock().unwrap();
    let id = path.into_inner();
    let len_before = notes.len();
    notes.retain(|n| n.id != id);

    if notes.len() < len_before {
        HttpResponse::Ok().json(ApiResponse::<()> {
            status: "success".to_string(),
            data: None,
            message: Some("Note deleted".to_string()),
        })
    } else {
        HttpResponse::NotFound().json(ApiResponse::<()> {
            status: "error".to_string(),
            data: None,
            message: Some("Note not found".to_string()),
        })
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .expect("PORT must be a number");

    let data = web::Data::new(AppState {
        notes: Mutex::new(vec![
            Note {
                id: Uuid::new_v4(),
                title: "Welcome".to_string(),
                content: "Welcome to the Note API!".to_string(),
                tags: vec!["welcome".to_string(), "demo".to_string()],
                created_at: Utc::now(),
            },
        ]),
    });

    log::info!("Note API server starting on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(data.clone())
            .wrap(middleware::Logger::default())
            .route("/health", web::get().to(health))
            .route("/notes", web::get().to(get_notes))
            .route("/notes/{id}", web::get().to(get_note))
            .route("/notes", web::post().to(create_note))
            .route("/notes/{id}", web::delete().to(delete_note))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
