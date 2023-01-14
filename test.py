from PIL import Image
import io
import pasteboard

pb = pasteboard.Pasteboard()

pb_image = pb.get_contents(pasteboard.String)
print(dir(pasteboard))
print(len(pb_image))
with open('test.png', 'wb') as f:
    f.write(pb_image)
# if pb_image:
#     print('here', len(pb_image))
#     with open('test.png' , 'wb') as f:
#         f.write(pb_image)
#     image = Image.open(io.BytesIO(pb_image))
#     if image.mode == 'RGBA':
#         image.load()
#         new_image = Image.new('RGB', image.size, (255, 255, 255))
#         new_image.paste(image, mask=image.split()[3])
#         image = new_image
#     data_bytes = io.BytesIO()
#     image.save(data_bytes, format='JPEG', quality=90)
#     data_bytes = data_bytes.getvalue()
#     pb.set_contents(data=data_bytes, type=pasteboard.TIFF)
#     print('Converted clipboard image to JPG.')
# else:
#     print('No image was copied.')
