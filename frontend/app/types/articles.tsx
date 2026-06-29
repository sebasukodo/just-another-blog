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
  Author: Author;
}
