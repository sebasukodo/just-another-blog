import type { Author } from "./users";

export interface Article {
  slug: string;
  title: string;
  description: string;
  body: string | null;
  tagList: string[];
  createdAt: string;
  updatedAt: string;
  favorited: boolean;
  favoritesCount: number;
  author: Author;
}

export interface Articles {
  articles: Article[];
  articlesCount: number;
}

export interface Tags {
  tags: string[];
}
