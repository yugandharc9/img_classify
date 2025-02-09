import {useState,useEffect} from 'react';
import './App.css';
import {SetImgDir,GetCategories,SetCategory,ClassifyImage} from "../wailsjs/go/main/App";

import { GetImageList } from "../wailsjs/go/main/App";

function App() {
    const [dir, setDir] = useState('');
    const [categories, setCategories] = useState([]);
    const updateDir = (result) => {
        setDir(result);
        fetchImages();
    }

        async function fetchImages() {
          const imageList = await GetImageList(); // Update the path
          setImages(imageList);
        }


    useEffect(() => { 
        GetCategories().then((res) => setCategories(res))
    },[categories]) 

    const [images, setImages] = useState([]);
    const [currentIndex, setCurrentIndex] = useState(0);

    function setImgDir() {
        SetImgDir().then(updateDir);
    }

      const showNextImage = () => {
        setCurrentIndex((prevIndex) => (prevIndex + 1) % images.length);
      };

    
      const classify = (category) => {  
          console.log(category,images[currentIndex]);
          ClassifyImage(category,images[currentIndex]).then((res)=> {console.log(res)})
          showNextImage();
      }
    
      const getCategoryToAdd = async () => {
        const enteredCategory = window.prompt("Enter category:");
        if (enteredCategory) {
            SetCategory(enteredCategory);
            setCategories((prevItems) => [...prevItems, enteredCategory]);

        }
      };

    return (
        <div className="container">
        {/* Permanent Sidebar */}
        { dir != "" &&
        <div className="sidebar">
          <h2>Categories</h2>
            <hr></hr>
          <p onClick={getCategoryToAdd}>+</p>
            <hr></hr>
            {categories.map((category, index) => (
            <p onClick={() => classify(category)} key={index}>{category}</p>
          ))}
        </div>  }
  
        {/* Main Content */}
        <div className="content">
          <div id="input" className="input-box">
            <button className="btn" style={{width: "30%"}} onClick={setImgDir}>Set Images Directory</button>
        
         { dir != "" &&  <p>Images Directory - {dir}</p>}
          </div>
  
          <div>
            {images.length > 0 ? (
              <>
                <img src={`http://localhost:5618/${images[currentIndex]}`} alt="Slideshow" width="600" />
                {/*<button onClick={showNextImage}>Next Image</button>*/}
              </>
            ) : (
              <p>Please set images directory</p>
            )}
          </div>
        </div>
      </div>
    );
    
}

export default App
