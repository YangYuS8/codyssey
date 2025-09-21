import {z} from 'zod';

// Problem
const nullableToOptional = <T extends z.ZodTypeAny>(schema: T) =>
  z.preprocess((v) => (v === null ? undefined : v), schema.optional());

export const ProblemSchema = z.object({
  id: z.string(),
  title: z.string(),
  description: nullableToOptional(z.string()),
  difficulty: nullableToOptional(z.string()),
  tags: nullableToOptional(z.array(z.string())),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});
export type Problem = z.infer<typeof ProblemSchema>;

// Submission (list item)
export const SubmissionItemSchema = z.object({
  id: z.string(),
  problemId: z.string(),
  userId: z.string(),
  status: z.string(),
  score: nullableToOptional(z.number()),
  language: nullableToOptional(z.string()),
  createdAt: z.string(),
  updatedAt: z.string().optional(),
  version: z.number().optional(),
});
export type SubmissionItem = z.infer<typeof SubmissionItemSchema>;

// JudgeRun (nested in submission detail)
export const JudgeRunSchema = z.object({
  id: z.string(),
  status: z.string(),
  durationMs: nullableToOptional(z.number()),
  createdAt: z.string(),
});

// Submission Detail
export const SubmissionDetailSchema = SubmissionItemSchema.extend({
  code: nullableToOptional(z.string()),
  runtimeMs: nullableToOptional(z.number()),
  memoryKb: nullableToOptional(z.number()),
  judgeRuns: nullableToOptional(z.array(JudgeRunSchema)),
});
export type SubmissionDetail = z.infer<typeof SubmissionDetailSchema>;

export function safeParseOrThrow<T>(schema: z.ZodType<T>, data: unknown): T {
  const r = schema.safeParse(data);
  if (!r.success) {
    // 抛出可追踪的结构化错误（后续可扩展专用错误类型）
    const e = new Error('SCHEMA_VALIDATION_FAILED');
    (e as any).issues = r.error.issues;
    throw e;
  }
  return r.data;
}
