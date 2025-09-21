"use client";
import React from 'react';
import { ToastProvider, ToastViewport } from './toast';

export const Toaster: React.FC = () => {
  return (
    <ToastProvider>
      <ToastViewport />
    </ToastProvider>
  );
};
