using ProductApi.Models;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSingleton<ProductStore>();

var app = builder.Build();

app.UseHttpsRedirection();

var products = app.MapGroup("/api/products");

products.MapGet("/", (ProductStore store) => Results.Ok(store.GetAll()));

products.MapGet("/{id:guid}", (Guid id, ProductStore store) =>
{
    var product = store.GetById(id);
    return product is not null ? Results.Ok(product) : Results.NotFound();
});

products.MapPost("/", (CreateProductRequest request, ProductStore store) =>
{
    var product = store.Add(request.Name, request.Price, request.Category);
    return Results.Created($"/api/products/{product.Id}", product);
});

products.MapPut("/{id:guid}", (Guid id, UpdateProductRequest request, ProductStore store) =>
{
    var product = store.Update(id, request.Name, request.Price, request.Category);
    return product is not null ? Results.Ok(product) : Results.NotFound();
});

products.MapDelete("/{id:guid}", (Guid id, ProductStore store) =>
{
    var deleted = store.Delete(id);
    return deleted ? Results.NoContent() : Results.NotFound();
});

app.MapGet("/health", () => Results.Ok(new { status = "healthy" }));

app.Run();
