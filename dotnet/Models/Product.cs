namespace ProductApi.Models;

public record Product(Guid Id, string Name, decimal Price, string Category, DateTime CreatedAt);

public record CreateProductRequest(string Name, decimal Price, string Category);

public record UpdateProductRequest(string Name, decimal Price, string Category);

public class ProductStore
{
    private readonly List<Product> _products = [];

    public List<Product> GetAll() => [.. _products];

    public Product? GetById(Guid id) => _products.Find(p => p.Id == id);

    public Product Add(string name, decimal price, string category)
    {
        var product = new Product(Guid.NewGuid(), name, price, category, DateTime.UtcNow);
        _products.Add(product);
        return product;
    }

    public Product? Update(Guid id, string name, decimal price, string category)
    {
        var index = _products.FindIndex(p => p.Id == id);
        if (index == -1) return null;

        var updated = _products[index] with { Name = name, Price = price, Category = category };
        _products[index] = updated;
        return updated;
    }

    public bool Delete(Guid id)
    {
        var product = _products.Find(p => p.Id == id);
        return product is not null && _products.Remove(product);
    }
}
