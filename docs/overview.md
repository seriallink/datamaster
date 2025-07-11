## Apresentação do Case

Para construir um projeto que fosse ao mesmo tempo realista, desafiador e reproduzível, optei por utilizar um conjunto de dados públicos e bem estruturados, sem a necessidade de depender de APIs, sistemas legados ou geradores de dados sintéticos.

O dataset escolhido foi o [BeerAdvocate Reviews](https://www.kaggle.com/datasets/thedevastator/1-5-million-beer-reviews-from-beer-advocate), publicado no Kaggle, contendo mais de **1.5 milhão de avaliações reais de cervejas**, incluindo notas de degustação, estilos, nomes de cervejarias, perfis de usuários e muito mais.

Para facilitar o uso desses dados no pipeline analítico, extraí e organizei os dados em quatro arquivos `.csv.gz`, disponíveis neste repositório: [github.com/seriallink/beer-datasets](https://github.com/seriallink/beer-datasets)

### Datasets disponíveis

| Arquivo          | Descrição                                            | Volume estimado   |
| ---------------- | ---------------------------------------------------- | ----------------- |
| `brewery.csv.gz` | Lista de cervejarias com identificador e nome        | 5.800+ linhas     |
| `beer.csv.gz`    | Detalhes das cervejas (estilo, teor alcoólico, etc.) | 66.000+ linhas    |
| `profile.csv.gz` | Perfis de usuários (sintéticos, criados via Faker)   | 33.000+ linhas    |
| `review.csv.gz`  | Avaliações completas com notas e comentários         | 1.500.000+ linhas |

> **Importante**: os e-mails em `profile.csv.gz` são totalmente sintéticos e foram gerados exclusivamente para fins de teste. Nenhuma informação pessoal foi utilizada, e os nomes de perfil são de domínio público.

---

[Voltar para a página inicial](../README.md)