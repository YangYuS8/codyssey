export type SupportedLanguage = 'cpp' | 'python' | 'go' | 'java';

const cpp = `#include <bits/stdc++.h>
using namespace std;

int main(){
    ios::sync_with_stdio(false);cin.tie(nullptr);
    // TODO: implement solution
    return 0;
}
`;

const python = `import sys

def solve():
    # TODO: implement solution
    pass

if __name__ == '__main__':
    solve()
`;

const go = `package main
import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    in := bufio.NewReader(os.Stdin)
    _ = in // TODO: implement solution
    fmt.Println("TODO")
}
`;

const java = `import java.io.*;
import java.util.*;

public class Main {
    public static void main(String[] args) throws Exception {
        FastScanner fs = new FastScanner(System.in);
        // TODO: implement solution
    }
    static class FastScanner {
        private final InputStream in; private final byte[] buffer = new byte[1<<16];
        private int ptr=0, len=0; FastScanner(InputStream is){in=is;}
        private int read() throws IOException { if (ptr>=len){ len=in.read(buffer); ptr=0; if(len<=0) return -1;} return buffer[ptr++]; }
        int nextInt() throws IOException { int c; while((c=read())<=32); int sign=1; if(c=='-'){sign=-1;c=read();} int x=c-'0'; while((c=read())>32){x=x*10+c-'0';} return x*sign; }
        String next() throws IOException { StringBuilder sb=new StringBuilder(); int c; while((c=read())<=32); while(c>32){ sb.append((char)c); c=read(); } return sb.toString(); }
    }
}
`;

const templates: Record<SupportedLanguage, string> = { cpp, python, go, java };

export function getTemplate(lang: SupportedLanguage): string {
  return templates[lang];
}

export function isSupportedLanguage(lang: string): lang is SupportedLanguage {
  return ['cpp','python','go','java'].includes(lang);
}
