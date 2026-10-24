using System.Security.Cryptography;
using System.Text;
using Neexitlag.Desktop;
using Xunit;

namespace NeExitLag.Tests;

public sealed class GameCatalogIntegrityTests
{
    static byte[] BundledBytes()
    {
        using var stream = typeof(GameCatalog).Assembly.GetManifestResourceStream(
            "Neexitlag.Desktop.Resources.GameCatalog.json");
        Assert.NotNull(stream);
        using var buffer = new MemoryStream();
        stream.CopyTo(buffer);
        return buffer.ToArray();
    }

    [Fact]
    public void EmbeddedCatalogHasLfAndExactPinnedHash()
    {
        var raw = BundledBytes();
        Assert.Contains((byte)'\n', raw);
        Assert.DoesNotContain((byte)'\r', raw);
        Assert.Equal(GameCatalog.BundledSha256, Convert.ToHexString(SHA256.HashData(raw)));
        Assert.NotEmpty(GameCatalog.Bundled().Games);
    }

    [Fact]
    public void CrLfConversionStillFailsIntegrityAndSignatureChecks()
    {
        var raw = BundledBytes();
        var crlf = Encoding.UTF8.GetBytes(Encoding.UTF8.GetString(raw).Replace("\n", "\r\n"));
        Assert.True(crlf.Length > raw.Length);
        var error = Assert.Throws<InvalidDataException>(() => GameCatalog.Parse(crlf, GameCatalog.BundledSha256));
        Assert.Equal("GAME_CATALOG_INTEGRITY_FAILED", error.Message);
        using var key = RSA.Create(2048);
        var signature = key.SignData(raw, HashAlgorithmName.SHA256, RSASignaturePadding.Pss);
        Assert.NotEmpty(GameCatalog.VerifySigned(raw, signature, key).Games);
        error = Assert.Throws<InvalidDataException>(() => GameCatalog.VerifySigned(crlf, signature, key));
        Assert.Equal("GAME_CATALOG_SIGNATURE_INVALID", error.Message);
    }
}
