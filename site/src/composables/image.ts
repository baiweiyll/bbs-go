export async function useUploadImage(file: File): Promise<string> {
    const formData = new FormData();
    formData.append("image", file, file.name);
    
    return useHttp("/bbsapi/upload", {
        method: "POST",
        body: formData,
    })
}